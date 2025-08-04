import React, {useEffect, useState} from 'react';
import {useTranslation} from 'react-i18next';
import {Card, Grid, Statistic, Segment, Label, Icon, Table, Message} from 'semantic-ui-react';
import {
  Bar,
  BarChart,
  CartesianGrid,
  Legend,
  Line,
  LineChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
  PieChart,
  Pie,
  Cell,
} from 'recharts';
import axios from 'axios';
import { API } from '../../helpers';
import './Dashboard.css';

// 在 Dashboard 组件内添加自定义配置
const chartConfig = {
  lineChart: {
    style: {
      background: '#fff',
      borderRadius: '8px',
    },
    line: {
      strokeWidth: 2,
      dot: false,
      activeDot: { r: 4 },
    },
    grid: {
      vertical: false,
      horizontal: true,
      opacity: 0.1,
    },
  },
  colors: {
    requests: '#4318FF',
    quota: '#00B5D8',
    tokens: '#6C63FF',
  },
  barColors: [
    '#4318FF', // 深紫色
    '#00B5D8', // 青色
    '#6C63FF', // 紫色
    '#05CD99', // 绿色
    '#FFB547', // 橙色
    '#FF5E7D', // 粉色
    '#41B883', // 翠绿
    '#7983FF', // 淡紫
    '#FF8F6B', // 珊瑚色
    '#49BEFF', // 天蓝
  ],
};

const Dashboard = () => {
  const { t } = useTranslation();
  const [data, setData] = useState([]);
  const [summaryData, setSummaryData] = useState({
    todayRequests: 0,
    todayQuota: 0,
    todayTokens: 0,
  });
  
  const [systemStats, setSystemStats] = useState({
    totalUsers: 0,
    totalTokens: 0,
    totalChannels: 0,
    activeChannels: 0,
    totalGroups: 0,
  });
  
  const [channelStats, setChannelStats] = useState([]);
  const [recentLogs, setRecentLogs] = useState([]);
  const [systemStatus, setSystemStatus] = useState({
    status: 'healthy',
    message: '系统运行正常',
  });

  useEffect(() => {
    fetchDashboardData();
    fetchSystemStats();
    fetchChannelStats();
    fetchRecentLogs();
  }, []);

  const fetchDashboardData = async () => {
    try {
      const response = await axios.get('/api/user/dashboard');
      if (response.data.success) {
        const dashboardData = response.data.data || [];
        setData(dashboardData);
        calculateSummary(dashboardData);
      }
    } catch (error) {
      console.error('Failed to fetch dashboard data:', error);
      setData([]);
      calculateSummary([]);
    }
  };

  const fetchSystemStats = async () => {
    try {
      const [usersRes, tokensRes, channelsRes, groupsRes] = await Promise.all([
        API.get('/api/user'),
        API.get('/api/token'),
        API.get('/api/channel'),
        API.get('/api/group/detail'),
      ]);

      const totalUsers = usersRes.data.success ? usersRes.data.data.length : 0;
      const totalTokens = tokensRes.data.success ? tokensRes.data.data.length : 0;
      const totalChannels = channelsRes.data.success ? channelsRes.data.data.length : 0;
      const activeChannels = channelsRes.data.success ? 
        channelsRes.data.data.filter(ch => ch.status === 1).length : 0;
      const totalGroups = groupsRes.data.success ? groupsRes.data.data.length : 0;

      setSystemStats({
        totalUsers,
        totalTokens,
        totalChannels,
        activeChannels,
        totalGroups,
      });
    } catch (error) {
      console.error('Failed to fetch system stats:', error);
    }
  };

  const fetchChannelStats = async () => {
    try {
      const response = await API.get('/api/channel');
      if (response.data.success) {
        const channels = response.data.data || [];
        const channelStatsData = channels.map(channel => ({
          name: channel.name || 'Unknown',
          type: channel.type,
          status: channel.status,
          balance: channel.balance || 0,
          responseTime: channel.response_time || 0,
        }));
        setChannelStats(channelStatsData);
      }
    } catch (error) {
      console.error('Failed to fetch channel stats:', error);
    }
  };

  const fetchRecentLogs = async () => {
    try {
      const response = await API.get('/api/log?p=1&size=10');
      if (response.data.success) {
        setRecentLogs(response.data.data || []);
      }
    } catch (error) {
      console.error('Failed to fetch recent logs:', error);
    }
  };

  const calculateSummary = (dashboardData) => {
    if (!Array.isArray(dashboardData) || dashboardData.length === 0) {
      setSummaryData({
        todayRequests: 0,
        todayQuota: 0,
        todayTokens: 0,
      });
      return;
    }

    const today = new Date().toISOString().split('T')[0];
    const todayData = dashboardData.filter((item) => item.Day === today);

    const summary = {
      todayRequests: todayData.reduce(
        (sum, item) => sum + item.RequestCount,
        0
      ),
      todayQuota:
        todayData.reduce((sum, item) => sum + item.Quota, 0) / 1000000,
      todayTokens: todayData.reduce(
        (sum, item) => sum + item.PromptTokens + item.CompletionTokens,
        0
      ),
    };

    setSummaryData(summary);
  };

  // 处理数据以供折线图使用，补充缺失的日期
  const processTimeSeriesData = () => {
    const dailyData = {};

    // 获取日期范围
    const dates = data.map((item) => item.Day);
    const maxDate = new Date(); // 总是使用今天作为最后一天
    let minDate =
      dates.length > 0
        ? new Date(Math.min(...dates.map((d) => new Date(d))))
        : new Date();

    // 确保至少显示7天的数据
    const sevenDaysAgo = new Date();
    sevenDaysAgo.setDate(sevenDaysAgo.getDate() - 6); // -6是因为包含今天
    if (minDate > sevenDaysAgo) {
      minDate = sevenDaysAgo;
    }

    // 生成所有日期
    for (let d = new Date(minDate); d <= maxDate; d.setDate(d.getDate() + 1)) {
      const dateStr = d.toISOString().split('T')[0];
      dailyData[dateStr] = {
        date: dateStr,
        requests: 0,
        quota: 0,
        tokens: 0,
      };
    }

    // 填充实际数据
    data.forEach((item) => {
      dailyData[item.Day].requests += item.RequestCount;
      dailyData[item.Day].quota += item.Quota / 1000000;
      dailyData[item.Day].tokens += item.PromptTokens + item.CompletionTokens;
    });

    return Object.values(dailyData).sort((a, b) =>
      a.date.localeCompare(b.date)
    );
  };

  // 处理数据以供堆叠柱状图使用
  const processModelData = () => {
    const timeData = {};

    // 获取日期范围
    const dates = data.map((item) => item.Day);
    const maxDate = new Date(); // 总是使用今天作为最后一天
    let minDate =
      dates.length > 0
        ? new Date(Math.min(...dates.map((d) => new Date(d))))
        : new Date();

    // 确保至少显示7天的数据
    const sevenDaysAgo = new Date();
    sevenDaysAgo.setDate(sevenDaysAgo.getDate() - 6); // -6是因为包含今天
    if (minDate > sevenDaysAgo) {
      minDate = sevenDaysAgo;
    }

    // 生成所有日期
    for (let d = new Date(minDate); d <= maxDate; d.setDate(d.getDate() + 1)) {
      const dateStr = d.toISOString().split('T')[0];
      timeData[dateStr] = {
        date: dateStr,
      };

      // 初始化所有模型的数据为0
      const models = [...new Set(data.map((item) => item.ModelName))];
      models.forEach((model) => {
        timeData[dateStr][model] = 0;
      });
    }

    // 填充实际数据
    data.forEach((item) => {
      timeData[item.Day][item.ModelName] =
        item.PromptTokens + item.CompletionTokens;
    });

    return Object.values(timeData).sort((a, b) => a.date.localeCompare(b.date));
  };

  // 获取所有唯一的模型名称
  const getUniqueModels = () => {
    return [...new Set(data.map((item) => item.ModelName))];
  };

  const timeSeriesData = processTimeSeriesData();
  const modelData = processModelData();
  const models = getUniqueModels();

  // 生成随机颜色
  const getRandomColor = (index) => {
    return chartConfig.barColors[index % chartConfig.barColors.length];
  };

  // 获取渠道状态标签
  const getChannelStatusLabel = (status) => {
    switch (status) {
      case 1:
        return <Label color="green">已启用</Label>;
      case 2:
        return <Label color="red">已禁用</Label>;
      default:
        return <Label color="grey">未知</Label>;
    }
  };

  // 获取渠道类型颜色
  const getChannelTypeColor = (type) => {
    const typeColors = {
      'openai': 'blue',
      'anthropic': 'purple',
      'google': 'red',
      'azure': 'teal',
      'baidu': 'orange',
      'ali': 'green',
      'default': 'grey',
    };
    return typeColors[type] || typeColors.default;
  };

  // 处理渠道数据用于饼图
  const processChannelDataForPie = () => {
    const typeCount = {};
    channelStats.forEach(channel => {
      typeCount[channel.type] = (typeCount[channel.type] || 0) + 1;
    });
    return Object.entries(typeCount).map(([type, count]) => ({
      name: type,
      value: count,
    }));
  };

  // 添加一个日期格式化函数
  const formatDate = (dateStr) => {
    const date = new Date(dateStr);
    return date.toLocaleDateString('zh-CN', {
      month: 'numeric',
      day: 'numeric',
    });
  };

  // 修改所有 XAxis 配置
  const xAxisConfig = {
    dataKey: 'date',
    axisLine: false,
    tickLine: false,
    tick: {
      fontSize: 12,
      fill: '#A3AED0',
      textAnchor: 'middle', // 文本居中对齐
    },
    tickFormatter: formatDate,
    interval: 0,
    minTickGap: 5,
    padding: { left: 30, right: 30 }, // 增加两侧的内边距，确保首尾标签完整显示
  };

  return (
    <div className='dashboard-container'>
      {/* 系统状态概览 */}
      <Grid columns={5} stackable style={{ marginBottom: '20px' }}>
        <Grid.Column>
          <Card fluid>
            <Card.Content textAlign="center">
              <Statistic>
                <Statistic.Value>
                  <Icon name="users" color="blue" />
                  {systemStats.totalUsers}
                </Statistic.Value>
                <Statistic.Label>总用户数</Statistic.Label>
              </Statistic>
            </Card.Content>
          </Card>
        </Grid.Column>
        <Grid.Column>
          <Card fluid>
            <Card.Content textAlign="center">
              <Statistic>
                <Statistic.Value>
                  <Icon name="key" color="green" />
                  {systemStats.totalTokens}
                </Statistic.Value>
                <Statistic.Label>总令牌数</Statistic.Label>
              </Statistic>
            </Card.Content>
          </Card>
        </Grid.Column>
        <Grid.Column>
          <Card fluid>
            <Card.Content textAlign="center">
              <Statistic>
                <Statistic.Value>
                  <Icon name="sitemap" color="purple" />
                  {systemStats.totalChannels}
                </Statistic.Value>
                <Statistic.Label>总渠道数</Statistic.Label>
              </Statistic>
            </Card.Content>
          </Card>
        </Grid.Column>
        <Grid.Column>
          <Card fluid>
            <Card.Content textAlign="center">
              <Statistic>
                <Statistic.Value>
                  <Icon name="check circle" color="teal" />
                  {systemStats.activeChannels}
                </Statistic.Value>
                <Statistic.Label>活跃渠道</Statistic.Label>
              </Statistic>
            </Card.Content>
          </Card>
        </Grid.Column>
        <Grid.Column>
          <Card fluid>
            <Card.Content textAlign="center">
              <Statistic>
                <Statistic.Value>
                  <Icon name="users" color="orange" />
                  {systemStats.totalGroups}
                </Statistic.Value>
                <Statistic.Label>用户组数</Statistic.Label>
              </Statistic>
            </Card.Content>
          </Card>
        </Grid.Column>
      </Grid>

      {/* 三个并排的折线图 */}
      <Grid columns={3} stackable className='charts-grid'>
        <Grid.Column>
          <Card fluid className='chart-card'>
            <Card.Content>
              <Card.Header>
                {t('dashboard.charts.requests.title')}
                {/* <span className='stat-value'>{summaryData.todayRequests}</span> */}
              </Card.Header>
              <div className='chart-container'>
                <ResponsiveContainer
                  width='100%'
                  height={120}
                  margin={{ left: 10, right: 10 }} // 调整容器边距
                >
                  <LineChart data={timeSeriesData}>
                    <CartesianGrid
                      strokeDasharray='3 3'
                      vertical={chartConfig.lineChart.grid.vertical}
                      horizontal={chartConfig.lineChart.grid.horizontal}
                      opacity={chartConfig.lineChart.grid.opacity}
                    />
                    <XAxis {...xAxisConfig} />
                    <YAxis hide={true} />
                    <Tooltip
                      contentStyle={{
                        background: '#fff',
                        border: 'none',
                        borderRadius: '4px',
                        boxShadow: '0 2px 8px rgba(0,0,0,0.1)',
                      }}
                      formatter={(value) => [
                        value,
                        t('dashboard.charts.requests.tooltip'),
                      ]}
                      labelFormatter={(label) =>
                        `${t(
                          'dashboard.statistics.tooltip.date'
                        )}: ${formatDate(label)}`
                      }
                    />
                    <Line
                      type='monotone'
                      dataKey='requests'
                      stroke={chartConfig.colors.requests}
                      strokeWidth={chartConfig.lineChart.line.strokeWidth}
                      dot={chartConfig.lineChart.line.dot}
                      activeDot={chartConfig.lineChart.line.activeDot}
                    />
                  </LineChart>
                </ResponsiveContainer>
              </div>
            </Card.Content>
          </Card>
        </Grid.Column>

        <Grid.Column>
          <Card fluid className='chart-card'>
            <Card.Content>
              <Card.Header>
                {t('dashboard.charts.quota.title')}
                {/* <span className='stat-value'>
                  ${summaryData.todayQuota.toFixed(3)}
                </span> */}
              </Card.Header>
              <div className='chart-container'>
                <ResponsiveContainer
                  width='100%'
                  height={120}
                  margin={{ left: 10, right: 10 }} // 调整容器边距
                >
                  <LineChart data={timeSeriesData}>
                    <CartesianGrid
                      strokeDasharray='3 3'
                      vertical={chartConfig.lineChart.grid.vertical}
                      horizontal={chartConfig.lineChart.grid.horizontal}
                      opacity={chartConfig.lineChart.grid.opacity}
                    />
                    <XAxis {...xAxisConfig} />
                    <YAxis hide={true} />
                    <Tooltip
                      contentStyle={{
                        background: '#fff',
                        border: 'none',
                        borderRadius: '4px',
                        boxShadow: '0 2px 8px rgba(0,0,0,0.1)',
                      }}
                      formatter={(value) => [
                        value.toFixed(6),
                        t('dashboard.charts.quota.tooltip'),
                      ]}
                      labelFormatter={(label) =>
                        `${t(
                          'dashboard.statistics.tooltip.date'
                        )}: ${formatDate(label)}`
                      }
                    />
                    <Line
                      type='monotone'
                      dataKey='quota'
                      stroke={chartConfig.colors.quota}
                      strokeWidth={chartConfig.lineChart.line.strokeWidth}
                      dot={chartConfig.lineChart.line.dot}
                      activeDot={chartConfig.lineChart.line.activeDot}
                    />
                  </LineChart>
                </ResponsiveContainer>
              </div>
            </Card.Content>
          </Card>
        </Grid.Column>

        <Grid.Column>
          <Card fluid className='chart-card'>
            <Card.Content>
              <Card.Header>
                {t('dashboard.charts.tokens.title')}
                {/* <span className='stat-value'>{summaryData.todayTokens}</span> */}
              </Card.Header>
              <div className='chart-container'>
                <ResponsiveContainer
                  width='100%'
                  height={120}
                  margin={{ left: 10, right: 10 }} // 调整容器边距
                >
                  <LineChart data={timeSeriesData}>
                    <CartesianGrid
                      strokeDasharray='3 3'
                      vertical={chartConfig.lineChart.grid.vertical}
                      horizontal={chartConfig.lineChart.grid.horizontal}
                      opacity={chartConfig.lineChart.grid.opacity}
                    />
                    <XAxis {...xAxisConfig} />
                    <YAxis hide={true} />
                    <Tooltip
                      contentStyle={{
                        background: '#fff',
                        border: 'none',
                        borderRadius: '4px',
                        boxShadow: '0 2px 8px rgba(0,0,0,0.1)',
                      }}
                      formatter={(value) => [
                        value,
                        t('dashboard.charts.tokens.tooltip'),
                      ]}
                      labelFormatter={(label) =>
                        `${t(
                          'dashboard.statistics.tooltip.date'
                        )}: ${formatDate(label)}`
                      }
                    />
                    <Line
                      type='monotone'
                      dataKey='tokens'
                      stroke={chartConfig.colors.tokens}
                      strokeWidth={chartConfig.lineChart.line.strokeWidth}
                      dot={chartConfig.lineChart.line.dot}
                      activeDot={chartConfig.lineChart.line.activeDot}
                    />
                  </LineChart>
                </ResponsiveContainer>
              </div>
            </Card.Content>
          </Card>
        </Grid.Column>
      </Grid>

      {/* 模型使用统计 */}
      <Card fluid className='chart-card'>
        <Card.Content>
          <Card.Header>{t('dashboard.statistics.title')}</Card.Header>
          <div className='chart-container'>
            <ResponsiveContainer width='100%' height={300}>
              <BarChart data={modelData}>
                <CartesianGrid
                  strokeDasharray='3 3'
                  vertical={false}
                  opacity={0.1}
                />
                <XAxis {...xAxisConfig} />
                <YAxis
                  axisLine={false}
                  tickLine={false}
                  tick={{ fontSize: 12, fill: '#A3AED0' }}
                />
                <Tooltip
                  contentStyle={{
                    background: '#fff',
                    border: 'none',
                    borderRadius: '4px',
                    boxShadow: '0 2px 8px rgba(0,0,0,0.1)',
                  }}
                  labelFormatter={(label) =>
                    `${t('dashboard.statistics.tooltip.date')}: ${formatDate(
                      label
                    )}`
                  }
                />
                <Legend
                  wrapperStyle={{
                    paddingTop: '20px',
                  }}
                />
                {models.map((model, index) => (
                  <Bar
                    key={model}
                    dataKey={model}
                    stackId='a'
                    fill={getRandomColor(index)}
                    name={model}
                    radius={[4, 4, 0, 0]}
                  />
                ))}
              </BarChart>
            </ResponsiveContainer>
          </div>
        </Card.Content>
      </Card>

      {/* 渠道统计和最近日志 */}
      <Grid columns={2} stackable style={{ marginTop: '20px' }}>
        <Grid.Column>
          <Card fluid>
            <Card.Content>
              <Card.Header>渠道类型分布</Card.Header>
              <div style={{ height: '300px' }}>
                <ResponsiveContainer width='100%' height='100%'>
                  <PieChart>
                    <Pie
                      data={processChannelDataForPie()}
                      cx='50%'
                      cy='50%'
                      labelLine={false}
                      label={({ name, percent }) => `${name} ${(percent * 100).toFixed(0)}%`}
                      outerRadius={80}
                      fill='#8884d8'
                      dataKey='value'
                    >
                      {processChannelDataForPie().map((entry, index) => (
                        <Cell key={`cell-${index}`} fill={getRandomColor(index)} />
                      ))}
                    </Pie>
                    <Tooltip />
                  </PieChart>
                </ResponsiveContainer>
              </div>
            </Card.Content>
          </Card>
        </Grid.Column>
        <Grid.Column>
          <Card fluid>
            <Card.Content>
              <Card.Header>最近系统日志</Card.Header>
              <div style={{ maxHeight: '300px', overflowY: 'auto' }}>
                {recentLogs.length > 0 ? (
                  recentLogs.map((log, index) => (
                    <Segment key={index} size='mini' style={{ margin: '5px 0' }}>
                      <div style={{ fontSize: '12px', color: '#666' }}>
                        {new Date(log.created_at).toLocaleString()}
                      </div>
                      <div style={{ fontSize: '14px', marginTop: '5px' }}>
                        {log.content}
                      </div>
                    </Segment>
                  ))
                ) : (
                  <Message info size='mini'>
                    暂无日志记录
                  </Message>
                )}
              </div>
            </Card.Content>
          </Card>
        </Grid.Column>
      </Grid>

      {/* 渠道状态表格 */}
      <Card fluid style={{ marginTop: '20px' }}>
        <Card.Content>
          <Card.Header>渠道状态概览</Card.Header>
          <Table celled compact>
            <Table.Header>
              <Table.Row>
                <Table.HeaderCell>渠道名称</Table.HeaderCell>
                <Table.HeaderCell>类型</Table.HeaderCell>
                <Table.HeaderCell>状态</Table.HeaderCell>
                <Table.HeaderCell>余额</Table.HeaderCell>
                <Table.HeaderCell>响应时间</Table.HeaderCell>
              </Table.Row>
            </Table.Header>
            <Table.Body>
              {channelStats.map((channel, index) => (
                <Table.Row key={index}>
                  <Table.Cell>{channel.name}</Table.Cell>
                  <Table.Cell>
                    <Label color={getChannelTypeColor(channel.type)}>
                      {channel.type}
                    </Label>
                  </Table.Cell>
                  <Table.Cell>{getChannelStatusLabel(channel.status)}</Table.Cell>
                  <Table.Cell>${channel.balance.toFixed(2)}</Table.Cell>
                  <Table.Cell>{channel.responseTime}ms</Table.Cell>
                </Table.Row>
              ))}
            </Table.Body>
          </Table>
        </Card.Content>
      </Card>
    </div>
  );
};

export default Dashboard;
