import React, { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Card, Grid, Label, Icon, Table, Message, Button, Dropdown } from 'semantic-ui-react';
import {
  CartesianGrid,
  Legend,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
  PieChart,
  Pie,
  Cell,
  Area,
  AreaChart,
  Sector,
} from 'recharts';
import { API } from '../../helpers';
import { CHANNEL_OPTIONS } from '../../constants/channel.constants';

// 科技感配色方案 - 专业配色，不超过三种主色
const techColorScheme = {
  primary: '#1E40AF',      // 深蓝色 - 主色调
  secondary: '#10B981',    // 翠绿色 - 成功/增长
  accent: '#F59E0B',       // 琥珀色 - 警告/重要
  background: '#F8FAFC',   // 浅灰背景
  surface: '#FFFFFF',      // 白色表面
  text: {
    primary: '#1F2937',    // 主要文字
    secondary: '#6B7280',  // 次要文字
    muted: '#9CA3AF',      // 弱化文字
  },
  chart: {
    primary: '#1E40AF',    // 主图表色
    secondary: '#10B981',  // 次图表色
    tertiary: '#F59E0B',   // 第三图表色
    success: '#059669',    // 成功色
    warning: '#D97706',    // 警告色
    error: '#DC2626',      // 错误色
  }
};

// 图表配置
const chartConfig = {
  lineChart: {
    style: {
      background: techColorScheme.surface,
      borderRadius: '12px',
    },
    line: {
      strokeWidth: 3,
      dot: false,
      activeDot: { r: 6, fill: techColorScheme.primary },
    },
    grid: {
      vertical: false,
      horizontal: true,
      opacity: 0.1,
    },
  },
  areaChart: {
    style: {
      background: techColorScheme.surface,
      borderRadius: '12px',
    },
    area: {
      opacity: 0.1,
    },
    line: {
      strokeWidth: 2,
      dot: false,
      activeDot: { r: 4 },
    },
  },
  colors: {
    requests: techColorScheme.chart.primary,
    quota: techColorScheme.chart.secondary,
    tokens: techColorScheme.chart.tertiary,
    success: techColorScheme.chart.success,
    error: techColorScheme.chart.error,
  },
  barColors: [
    techColorScheme.chart.primary,
    techColorScheme.chart.secondary,
    techColorScheme.chart.tertiary,
    techColorScheme.chart.success,
    techColorScheme.chart.warning,
    techColorScheme.chart.error,
  ],
};

const Dashboard = () => {
  useTranslation(); // 保留国际化支持
  const [data, setData] = useState([]);
  const [systemStats, setSystemStats] = useState({
    totalUsers: 0,
    totalTokens: 0,
    totalChannels: 0,
    activeChannels: 0,
    totalGroups: 0,
  });
  
  const [channelStats, setChannelStats] = useState([]);
  const [timeRange, setTimeRange] = useState('7d'); // API趋势图的时间范围
  const [channelTimeRange, setChannelTimeRange] = useState('7d'); // 渠道使用分布的时间范围
  const [loading, setLoading] = useState(true);
  const [apiDataKey, setApiDataKey] = useState(Date.now()); // 用于强制刷新API趋势图
  const [channelDataKey, setChannelDataKey] = useState(Date.now()); // 用于强制刷新渠道图表
  const [activeIndex, setActiveIndex] = useState(0); // 用于饼图交互

  // API趋势图数据加载
  useEffect(() => {
    const fetchApiData = async () => {
      await fetchDashboardData();
      fetchSystemStats();
    };
    
    fetchApiData();
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [timeRange]);
  
  // 渠道使用分布数据加载
  useEffect(() => {
    fetchChannelStats();
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [channelTimeRange]);

  const fetchDashboardData = async () => {
    try {
      // 根据时间范围计算开始和结束日期
      const now = new Date();
      let startDate;
      
      switch(timeRange) {
        case '30d':
          startDate = new Date(now);
          startDate.setDate(now.getDate() - 30);
          break;
        case '90d':
          startDate = new Date(now);
          startDate.setDate(now.getDate() - 90);
          break;
        case '7d':
        default:
          startDate = new Date(now);
          startDate.setDate(now.getDate() - 7);
          break;
      }
      
      // 将日期转换为Unix时间戳
      // const startTimestamp = Math.floor(startDate.getTime() / 1000);
      // const endTimestamp = Math.floor(now.getTime() / 1000);
      
      // 请求指定时间范围的数据
      const response = await API.get(`/api/user/dashboard?start=${startDate.getTime() / 1000}&end=${now.getTime() / 1000}`);
      
      if (response.data.success) {
        const dashboardData = response.data.data || [];
        setData(dashboardData);
        handleDataLoaded(dashboardData);
      }
    } catch (error) {
      console.error('Failed to fetch dashboard data:', error);
      setData([]);
      handleDataLoaded([]);
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
      // 根据时间范围计算开始和结束日期
      const now = new Date();
      let startDate;
      
      switch(channelTimeRange) {
        case '30d':
          startDate = new Date(now);
          startDate.setDate(now.getDate() - 30);
          break;
        case '90d':
          startDate = new Date(now);
          startDate.setDate(now.getDate() - 90);
          break;
        case '7d':
        default:
          startDate = new Date(now);
          startDate.setDate(now.getDate() - 7);
          break;
      }
      
      // 将日期转换为Unix时间戳
      // const startTimestamp = Math.floor(startDate.getTime() / 1000);
      // const endTimestamp = Math.floor(now.getTime() / 1000);
      
      // 请求渠道数据，不添加时间范围参数，因为渠道列表不需要时间过滤
      const response = await API.get('/api/channel');
      
      if (response.data.success) {
        const channels = response.data.data || [];
        
        const channelStatsData = channels.map(channel => ({
          id: channel.id,
          name: channel.name || 'Unknown',
          type: channel.type,
          status: channel.status,
          responseTime: channel.response_time || 0, // 后端字段名是 response_time，单位是毫秒
          requestCount: channel.used_quota || 0, // 后端字段名是 used_quota
        }));
        
        setChannelStats(channelStatsData);
      }
    } catch (error) {
      console.error('Failed to fetch channel stats:', error);
    }
  };

  // 设置加载状态完成
  const finishLoading = () => {
    setLoading(false);
  };

  // 处理数据加载完成
  const handleDataLoaded = (dashboardData) => {
    if (Array.isArray(dashboardData) && dashboardData.length > 0) {
      // 数据加载成功
      finishLoading();
    } else {
      // 无数据
      finishLoading();
    }
  };

  // 处理API请求趋势数据
  const processApiRequestData = () => {
    if (!Array.isArray(data) || data.length === 0) {
      return [];
    }

    // 按日期分组数据
    const groupedData = {};
    data.forEach(item => {
      const date = item.Day;
      if (!groupedData[date]) {
        groupedData[date] = {
          date,
          requests: 0,
          tokens: 0,
          quota: 0,
        };
      }
      groupedData[date].requests += item.RequestCount || 0;
      groupedData[date].tokens += (item.PromptTokens || 0) + (item.CompletionTokens || 0);
      groupedData[date].quota += item.Quota || 0;
    });

    return Object.values(groupedData).sort((a, b) => new Date(a.date) - new Date(b.date));
  };

  // 处理渠道使用分布数据
  const processChannelUsageData = () => {
    if (!Array.isArray(channelStats) || channelStats.length === 0) {
      return [];
    }

    // 按渠道名称统计请求量
    const channelData = channelStats
      .filter(channel => channel.status === 1 && channel.requestCount > 0) // 只统计活跃且有请求的渠道
      .map(channel => ({
        name: channel.name,
        value: channel.requestCount
      }))
      .sort((a, b) => b.value - a.value);
    
    // 取前7个渠道，其余归为"其他"类别
    if (channelData.length > 7) {
      const topChannels = channelData.slice(0, 7);
      const otherChannels = channelData.slice(7);
      const otherValue = otherChannels.reduce((sum, item) => sum + item.value, 0);
      
      if (otherValue > 0) {
        topChannels.push({
          name: '其他渠道',
          value: otherValue
        });
      }
      
      return topChannels;
    }
    
    return channelData;
  };



  // 获取渠道类型名称
  const getChannelTypeName = (typeId) => {
    const channelOption = CHANNEL_OPTIONS.find(option => option.value === typeId);
    return channelOption ? channelOption.text : `类型${typeId}`;
  };

  // 获取渠道类型颜色
  const getChannelTypeColor = (typeId) => {
    const channelOption = CHANNEL_OPTIONS.find(option => option.value === typeId);
    if (!channelOption) return techColorScheme.text.muted;
    
    const colorMap = {
      'green': techColorScheme.chart.success,
      'blue': techColorScheme.chart.primary,
      'orange': techColorScheme.chart.warning,
      'red': techColorScheme.chart.error,
      'purple': techColorScheme.chart.tertiary,
      'teal': techColorScheme.chart.secondary,
      'black': techColorScheme.text.primary,
      'olive': techColorScheme.chart.success,
      'brown': techColorScheme.chart.warning,
      'violet': techColorScheme.chart.tertiary,
      'pink': techColorScheme.chart.error,
    };
    
    return colorMap[channelOption.color] || techColorScheme.text.muted;
  };

  // 获取渠道状态标签
  const getChannelStatusLabel = (status) => {
    switch (status) {
      case 1:
        return <Label color="green" size="tiny">已启用</Label>;
      case 2:
        return <Label color="red" size="tiny">已禁用</Label>;
      case 3:
        return <Label color="orange" size="tiny">自动禁用</Label>;
      default:
        return <Label color="grey" size="tiny">未知</Label>;
    }
  };

  // 格式化响应时间
  const formatResponseTime = (time) => {
    if (time < 1000) return `${time}ms`;
    return `${(time / 1000).toFixed(1)}s`;
  };

  // 格式化日期
  const formatDate = (dateStr) => {
    const date = new Date(dateStr);
    return date.toLocaleDateString('zh-CN', {
      month: 'numeric',
      day: 'numeric',
    });
  };
  
  // 处理饼图点击事件
  const onPieEnter = (_, index) => {
    setActiveIndex(index);
  };
  
  // 自定义活动形状组件
  const renderActiveShape = (props) => {
    const RADIAN = Math.PI / 180;
    const { 
      cx, cy, midAngle, innerRadius, outerRadius, startAngle, endAngle,
      fill, payload, percent, value
    } = props;
    
    const sin = Math.sin(-RADIAN * midAngle);
    const cos = Math.cos(-RADIAN * midAngle);
    const sx = cx + (outerRadius + 10) * cos;
    const sy = cy + (outerRadius + 10) * sin;
    const mx = cx + (outerRadius + 30) * cos;
    const my = cy + (outerRadius + 30) * sin;
    const ex = mx + (cos >= 0 ? 1 : -1) * 22;
    const ey = my;
    const textAnchor = cos >= 0 ? 'start' : 'end';
  
    return (
      <g>
        <text x={cx} y={cy} dy={8} textAnchor="middle" fill={fill}>
          {payload.name}
        </text>
        <Sector
          cx={cx}
          cy={cy}
          innerRadius={innerRadius}
          outerRadius={outerRadius + 10}
          startAngle={startAngle}
          endAngle={endAngle}
          fill={fill}
        />
        <Sector
          cx={cx}
          cy={cy}
          startAngle={startAngle}
          endAngle={endAngle}
          innerRadius={outerRadius + 6}
          outerRadius={outerRadius + 10}
          fill={fill}
        />
        <path d={`M${sx},${sy}L${mx},${my}L${ex},${ey}`} stroke={fill} fill="none" />
        <circle cx={ex} cy={ey} r={2} fill={fill} stroke="none" />
        <text x={ex + (cos >= 0 ? 1 : -1) * 12} y={ey} textAnchor={textAnchor} fill="#333">{`${payload.name}`}</text>
        <text x={ex + (cos >= 0 ? 1 : -1) * 12} y={ey} dy={18} textAnchor={textAnchor} fill="#999">
          {`${value} 次 (${(percent * 100).toFixed(2)}%)`}
        </text>
      </g>
    );
  };

  // 时间范围选项
  const timeRangeOptions = [
    { key: '7d', text: '最近7天', value: '7d' },
    { key: '30d', text: '最近30天', value: '30d' },
    { key: '90d', text: '最近90天', value: '90d' },
  ];

  const timeSeriesData = processApiRequestData();
  const channelUsageData = processChannelUsageData();

  if (loading) {
    return (
      <div style={{
        padding: '24px',
        backgroundColor: techColorScheme.background,
        minHeight: '100vh',
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center'
      }}>
        <Message info>
          <Message.Header>加载中...</Message.Header>
          <p>正在获取仪表板数据</p>
        </Message>
      </div>
    );
  }

  return (
    <div style={{
      padding: '24px',
      backgroundColor: techColorScheme.background,
      minHeight: '100vh',
      maxWidth: '1400px',
      margin: '0 auto'
    }}>
      {/* 仪表盘标题 */}
      <div style={{
        marginBottom: '24px'
      }}>
        <h2 style={{
          color: techColorScheme.text.primary,
          fontSize: '24px',
          fontWeight: '600',
          margin: '0'
        }}>
          仪表盘数据概览
        </h2>
      </div>
      
      {/* 顶部统计卡片 - 可点击跳转 */}
      <Grid columns={4} stackable style={{ marginBottom: '32px' }}>
        <Grid.Column>
          <Card 
            fluid 
            style={{
              background: `linear-gradient(135deg, ${techColorScheme.primary} 0%, #3B82F6 100%)`,
              borderRadius: '16px',
              border: 'none',
              boxShadow: '0 4px 20px rgba(30, 64, 175, 0.15)',
              cursor: 'pointer',
              transition: 'transform 0.2s ease, box-shadow 0.2s ease',
            }}
            onClick={() => window.location.href = '/user'}
            onMouseEnter={(e) => {
              e.currentTarget.style.transform = 'translateY(-2px)';
              e.currentTarget.style.boxShadow = '0 8px 25px rgba(30, 64, 175, 0.25)';
            }}
            onMouseLeave={(e) => {
              e.currentTarget.style.transform = 'translateY(0)';
              e.currentTarget.style.boxShadow = '0 4px 20px rgba(30, 64, 175, 0.15)';
            }}
          >
            <Card.Content textAlign="center" style={{ padding: '24px 16px' }}>
              <div style={{ marginBottom: '12px' }}>
                <Icon name="users" size="large" style={{ color: 'rgba(255,255,255,0.9)' }} />
              </div>
              <div style={{ color: 'white', fontSize: '28px', fontWeight: '700', marginBottom: '4px' }}>
                {systemStats.totalUsers.toLocaleString()}
              </div>
              <div style={{ color: 'rgba(255,255,255,0.8)', fontSize: '14px', fontWeight: '500' }}>
                总用户数
              </div>
            </Card.Content>
          </Card>
        </Grid.Column>
        
        <Grid.Column>
          <Card 
            fluid 
            style={{
              background: `linear-gradient(135deg, ${techColorScheme.secondary} 0%, #34D399 100%)`,
              borderRadius: '16px',
              border: 'none',
              boxShadow: '0 4px 20px rgba(16, 185, 129, 0.15)',
              cursor: 'pointer',
              transition: 'transform 0.2s ease, box-shadow 0.2s ease',
            }}
            onClick={() => window.location.href = '/token'}
            onMouseEnter={(e) => {
              e.currentTarget.style.transform = 'translateY(-2px)';
              e.currentTarget.style.boxShadow = '0 8px 25px rgba(16, 185, 129, 0.25)';
            }}
            onMouseLeave={(e) => {
              e.currentTarget.style.transform = 'translateY(0)';
              e.currentTarget.style.boxShadow = '0 4px 20px rgba(16, 185, 129, 0.15)';
            }}
          >
            <Card.Content textAlign="center" style={{ padding: '24px 16px' }}>
              <div style={{ marginBottom: '12px' }}>
                <Icon name="key" size="large" style={{ color: 'rgba(255,255,255,0.9)' }} />
              </div>
              <div style={{ color: 'white', fontSize: '28px', fontWeight: '700', marginBottom: '4px' }}>
                {systemStats.totalTokens.toLocaleString()}
              </div>
              <div style={{ color: 'rgba(255,255,255,0.8)', fontSize: '14px', fontWeight: '500' }}>
                总令牌数
              </div>
            </Card.Content>
          </Card>
        </Grid.Column>
        
        <Grid.Column>
          <Card 
            fluid 
            style={{
              background: `linear-gradient(135deg, ${techColorScheme.accent} 0%, #FBBF24 100%)`,
              borderRadius: '16px',
              border: 'none',
              boxShadow: '0 4px 20px rgba(245, 158, 11, 0.15)',
              cursor: 'pointer',
              transition: 'transform 0.2s ease, box-shadow 0.2s ease',
            }}
            onClick={() => window.location.href = '/channel'}
            onMouseEnter={(e) => {
              e.currentTarget.style.transform = 'translateY(-2px)';
              e.currentTarget.style.boxShadow = '0 8px 25px rgba(245, 158, 11, 0.25)';
            }}
            onMouseLeave={(e) => {
              e.currentTarget.style.transform = 'translateY(0)';
              e.currentTarget.style.boxShadow = '0 4px 20px rgba(245, 158, 11, 0.15)';
            }}
          >
            <Card.Content textAlign="center" style={{ padding: '24px 16px' }}>
              <div style={{ marginBottom: '12px' }}>
                <Icon name="sitemap" size="large" style={{ color: 'rgba(255,255,255,0.9)' }} />
              </div>
              <div style={{ color: 'white', fontSize: '28px', fontWeight: '700', marginBottom: '4px' }}>
                {systemStats.totalChannels}
              </div>
              <div style={{ color: 'rgba(255,255,255,0.8)', fontSize: '14px', fontWeight: '500' }}>
                总渠道数
              </div>
            </Card.Content>
          </Card>
        </Grid.Column>
        
        <Grid.Column>
          <Card 
            fluid 
            style={{
              background: `linear-gradient(135deg, ${techColorScheme.chart.warning} 0%, #F59E0B 100%)`,
              borderRadius: '16px',
              border: 'none',
              boxShadow: '0 4px 20px rgba(217, 119, 6, 0.15)',
              cursor: 'pointer',
              transition: 'transform 0.2s ease, box-shadow 0.2s ease',
            }}
            onClick={() => window.location.href = '/group'}
            onMouseEnter={(e) => {
              e.currentTarget.style.transform = 'translateY(-2px)';
              e.currentTarget.style.boxShadow = '0 8px 25px rgba(217, 119, 6, 0.25)';
            }}
            onMouseLeave={(e) => {
              e.currentTarget.style.transform = 'translateY(0)';
              e.currentTarget.style.boxShadow = '0 4px 20px rgba(217, 119, 6, 0.15)';
            }}
          >
            <Card.Content textAlign="center" style={{ padding: '24px 16px' }}>
              <div style={{ marginBottom: '12px' }}>
                <Icon name="users" size="large" style={{ color: 'rgba(255,255,255,0.9)' }} />
              </div>
              <div style={{ color: 'white', fontSize: '28px', fontWeight: '700', marginBottom: '4px' }}>
                {systemStats.totalGroups}
              </div>
              <div style={{ color: 'rgba(255,255,255,0.8)', fontSize: '14px', fontWeight: '500' }}>
                用户组数
              </div>
            </Card.Content>
          </Card>
        </Grid.Column>
      </Grid>

      {/* API请求趋势图 - 使用真实数据 */}
      <Card fluid style={{
        marginBottom: '24px',
        borderRadius: '16px',
        border: 'none',
        boxShadow: '0 2px 12px rgba(0, 0, 0, 0.04)',
      }}>
        <Card.Content style={{ padding: '24px' }}>
          <div style={{
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            marginBottom: '24px'
          }}>
            <div>
              <h3 style={{
                color: techColorScheme.text.primary,
                fontSize: '20px',
                fontWeight: '600',
                margin: '0 0 4px 0'
              }}>
                API请求趋势
              </h3>
              <p style={{
                color: techColorScheme.text.secondary,
                fontSize: '14px',
                margin: '0'
              }}>
                监控OneAPI系统的请求量、令牌使用量和配额消耗
              </p>
            </div>
            <div style={{
              display: 'flex',
              alignItems: 'center'
            }}>
              <span style={{
                color: techColorScheme.text.secondary,
                marginRight: '8px',
                fontSize: '14px'
              }}>
                时间范围:
              </span>
              <Dropdown
                value={timeRange}
                options={timeRangeOptions}
                onChange={(e, { value }) => {
                  setTimeRange(value);
                  setApiDataKey(Date.now()); // 强制刷新API趋势图
                }}
                style={{ minWidth: '120px' }}
              />
            </div>
          </div>
          
          {timeSeriesData.length > 0 ? (
            <ResponsiveContainer width="100%" height={400} key={`api-chart-${apiDataKey}`}>
              <AreaChart data={timeSeriesData}>
                <defs>
                  <linearGradient id="requestsGradient" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor={techColorScheme.chart.primary} stopOpacity={0.3}/>
                    <stop offset="95%" stopColor={techColorScheme.chart.primary} stopOpacity={0}/>
                  </linearGradient>
                  <linearGradient id="tokensGradient" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor={techColorScheme.chart.secondary} stopOpacity={0.3}/>
                    <stop offset="95%" stopColor={techColorScheme.chart.secondary} stopOpacity={0}/>
                  </linearGradient>
                </defs>
                <CartesianGrid strokeDasharray="3 3" opacity={0.1} />
                <XAxis 
                  dataKey="date" 
                  tickFormatter={formatDate}
                  tick={{ fontSize: 12, fill: techColorScheme.text.secondary }}
                  axisLine={false}
                  tickLine={false}
                />
                <YAxis 
                  yAxisId="left"
                  tick={{ fontSize: 12, fill: techColorScheme.text.secondary }}
                  axisLine={false}
                  tickLine={false}
                />
                <YAxis 
                  yAxisId="right" 
                  orientation="right"
                  tick={{ fontSize: 12, fill: techColorScheme.text.secondary }}
                  axisLine={false}
                  tickLine={false}
                />
                <Tooltip
                  contentStyle={{
                    background: techColorScheme.surface,
                    border: 'none',
                    borderRadius: '8px',
                    boxShadow: '0 4px 20px rgba(0,0,0,0.1)',
                  }}
                  labelFormatter={(label) => `日期: ${formatDate(label)}`}
                  formatter={(value, name) => {
                    if (name === "请求量 (次)") {
                      return [`${value} 次`, name];
                    } else if (name === "令牌数 (个)") {
                      return [`${value} 个`, name];
                    }
                    return [value, name];
                  }}
                />
                <Legend />
                <Area
                  yAxisId="left"
                  type="monotone"
                  dataKey="requests"
                  stroke={techColorScheme.chart.primary}
                  fill="url(#requestsGradient)"
                  name="请求量 (次)"
                  strokeWidth={2}
                />
                <Area
                  yAxisId="right"
                  type="monotone"
                  dataKey="tokens"
                  stroke={techColorScheme.chart.secondary}
                  fill="url(#tokensGradient)"
                  name="令牌数 (个)"
                  strokeWidth={2}
                />
              </AreaChart>
            </ResponsiveContainer>
          ) : (
            <div style={{
              height: '400px',
              display: 'flex',
              flexDirection: 'column',
              justifyContent: 'center',
              alignItems: 'center',
              color: techColorScheme.text.secondary,
              fontSize: '16px',
              textAlign: 'center'
            }}>
              <Icon name="line chart" size="large" style={{ marginBottom: '12px', opacity: 0.5 }} />
              <div>暂无API请求数据</div>
              <div style={{ fontSize: '14px', marginTop: '8px', opacity: 0.7 }}>
                请确保在选定时间范围内有API请求记录
              </div>
            </div>
          )}
        </Card.Content>
      </Card>

      {/* 渠道使用分布 */}
      <Grid columns={1} stackable style={{ marginBottom: '24px' }}>
        <Grid.Column>
          <Card fluid style={{
            borderRadius: '16px',
            border: 'none',
            boxShadow: '0 2px 12px rgba(0, 0, 0, 0.04)',
            height: '100%'
          }}>
            <Card.Content style={{ padding: '24px' }}>
              <div style={{
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'flex-start',
                marginBottom: '16px'
              }}>
                <div>
                  <h3 style={{
                    color: techColorScheme.text.primary,
                    fontSize: '18px',
                    fontWeight: '600',
                    margin: '0 0 4px 0'
                  }}>
                    渠道使用分布
                  </h3>
                  <p style={{
                    color: techColorScheme.text.secondary,
                    fontSize: '14px',
                    margin: '0'
                  }}>
                    各渠道的请求量分布情况
                  </p>
                </div>
                <div style={{
                  display: 'flex',
                  alignItems: 'center'
                }}>
                  <span style={{
                    color: techColorScheme.text.secondary,
                    marginRight: '8px',
                    fontSize: '14px'
                  }}>
                    时间范围:
                  </span>
                  <Dropdown
                    value={channelTimeRange}
                    options={timeRangeOptions}
                    onChange={(e, { value }) => {
                      setChannelTimeRange(value);
                      setChannelDataKey(Date.now()); // 强制刷新渠道图表
                    }}
                    style={{ minWidth: '120px' }}
                  />
                </div>
              </div>
              {channelUsageData.length > 0 ? (
                <ResponsiveContainer width="100%" height={300} key={`channel-chart-${channelDataKey}`}>
                  <PieChart>
                    <Pie
                      activeIndex={activeIndex}
                      activeShape={renderActiveShape}
                      data={channelUsageData}
                      cx="50%"
                      cy="50%"
                      labelLine={false}
                      label={false}
                      outerRadius={80}
                      innerRadius={40}
                      fill="#8884d8"
                      dataKey="value"
                      paddingAngle={3}
                      onMouseEnter={onPieEnter}
                      onClick={onPieEnter}
                    >
                      {channelUsageData.map((entry, index) => (
                        <Cell key={`cell-${index}`} fill={chartConfig.barColors[index % chartConfig.barColors.length]} />
                      ))}
                    </Pie>
                    <Tooltip
                      contentStyle={{
                        background: techColorScheme.surface,
                        border: 'none',
                        borderRadius: '8px',
                        boxShadow: '0 4px 20px rgba(0,0,0,0.1)',
                      }}
                      formatter={(value, name) => [`${value} 次 (${((value / channelUsageData.reduce((sum, item) => sum + item.value, 0)) * 100).toFixed(1)}%)`, name]}
                    />
                    <Legend
                      layout="vertical"
                      verticalAlign="middle"
                      align="right"
                      wrapperStyle={{ fontSize: '12px', paddingLeft: '10px' }}
                    />
                  </PieChart>
                </ResponsiveContainer>
              ) : (
                <div style={{
                  height: '300px',
                  display: 'flex',
                  flexDirection: 'column',
                  justifyContent: 'center',
                  alignItems: 'center',
                  color: techColorScheme.text.secondary,
                  fontSize: '16px',
                  textAlign: 'center'
                }}>
                  <Icon name="pie chart" size="large" style={{ marginBottom: '12px', opacity: 0.5 }} />
                  <div>暂无渠道使用数据</div>
                  <div style={{ fontSize: '14px', marginTop: '8px', opacity: 0.7 }}>
                    请确保有启用的渠道且有请求数据
                  </div>
                </div>
              )}
            </Card.Content>
          </Card>
        </Grid.Column>
      </Grid>

      {/* 渠道状态概览 - 优化表格显示 */}
      <Card fluid style={{
        borderRadius: '16px',
        border: 'none',
        boxShadow: '0 2px 12px rgba(0, 0, 0, 0.04)',
      }}>
        <Card.Content style={{ padding: '24px' }}>
          <div style={{
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            marginBottom: '20px'
          }}>
            <div>
              <h3 style={{
                color: techColorScheme.text.primary,
                fontSize: '18px',
                fontWeight: '600',
                margin: '0 0 4px 0'
              }}>
                渠道状态概览
              </h3>
              <p style={{
                color: techColorScheme.text.secondary,
                fontSize: '14px',
                margin: '0'
              }}>
                监控各渠道的运行状态、响应时间和请求统计
              </p>
            </div>
            <Button 
              primary 
              size="small"
              onClick={() => window.location.href = '/channel'}
              style={{
                background: techColorScheme.primary,
                borderRadius: '8px'
              }}
            >
              管理渠道
            </Button>
          </div>
          
          {channelStats.length > 0 ? (
            <>
              <Table celled compact style={{ marginTop: '16px' }}>
                <Table.Header>
                  <Table.Row>
                    <Table.HeaderCell style={{ background: techColorScheme.background }}>渠道名称</Table.HeaderCell>
                    <Table.HeaderCell style={{ background: techColorScheme.background }}>渠道类型</Table.HeaderCell>
                    <Table.HeaderCell style={{ background: techColorScheme.background }}>运行状态</Table.HeaderCell>
                    <Table.HeaderCell style={{ background: techColorScheme.background }}>响应时间</Table.HeaderCell>
                    <Table.HeaderCell style={{ background: techColorScheme.background }}>请求量</Table.HeaderCell>
                  </Table.Row>
                </Table.Header>
                <Table.Body>
                  {channelStats.slice(0, 10).map((channel, index) => (
                    <Table.Row key={index}>
                      <Table.Cell style={{ fontWeight: '500' }}>{channel.name}</Table.Cell>
                      <Table.Cell>
                        <Label 
                          size="tiny" 
                          style={{
                            background: getChannelTypeColor(channel.type),
                            color: 'white',
                            borderRadius: '4px'
                          }}
                        >
                          {getChannelTypeName(channel.type)}
                        </Label>
                      </Table.Cell>
                      <Table.Cell>{getChannelStatusLabel(channel.status)}</Table.Cell>
                      <Table.Cell>
                        <span style={{
                          color: channel.responseTime < 1000 ? techColorScheme.chart.success : 
                                 channel.responseTime < 3000 ? techColorScheme.chart.warning : 
                                 techColorScheme.chart.error,
                          fontWeight: '500'
                        }}>
                          {formatResponseTime(channel.responseTime)}
                        </span>
                      </Table.Cell>
                      <Table.Cell>
                        <span style={{ fontWeight: '500' }}>
                          {channel.requestCount.toLocaleString()}
                        </span>
                      </Table.Cell>
                    </Table.Row>
                  ))}
                </Table.Body>
              </Table>
              
              {channelStats.length > 10 && (
                <div style={{
                  textAlign: 'center',
                  marginTop: '16px',
                  padding: '12px',
                  color: techColorScheme.text.secondary,
                  fontSize: '14px'
                }}>
                  显示前10个渠道，共{channelStats.length}个渠道
                  <Button 
                    basic 
                    size="small" 
                    style={{ marginLeft: '12px' }}
                    onClick={() => window.location.href = '/channel'}
                  >
                    查看全部
                  </Button>
                </div>
              )}
            </>
          ) : (
            <div style={{
              height: '200px',
              display: 'flex',
              flexDirection: 'column',
              justifyContent: 'center',
              alignItems: 'center',
              color: techColorScheme.text.secondary,
              fontSize: '16px',
              textAlign: 'center'
            }}>
              <Icon name="table" size="large" style={{ marginBottom: '12px', opacity: 0.5 }} />
              <div>暂无渠道数据</div>
              <div style={{ fontSize: '14px', marginTop: '8px', opacity: 0.7 }}>
                请先添加渠道或确保渠道数据已加载
              </div>
            </div>
          )}
        </Card.Content>
      </Card>
    </div>
  );
};

export default Dashboard;
