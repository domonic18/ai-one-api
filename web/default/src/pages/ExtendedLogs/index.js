import React, { useEffect, useState, useCallback } from 'react';
import { useTranslation } from 'react-i18next';
import {
  Button,
  Card,
  Form,
  Header,
  Label,
  Pagination,
  Segment,
  Table,
  Modal,
  Grid,
  Icon,
  Message,
  Dropdown,
  Input,
  Popup,
} from 'semantic-ui-react';
import {
  API,
  copy,
  isAdmin,
  showError,
  showSuccess,
  showWarning,
  timestamp2string,
} from '../../helpers';
import { ITEMS_PER_PAGE } from '../../constants';
import { renderColorLabel, renderQuota } from '../../helpers/render';

const ExtendedLogs = () => {
  const { t } = useTranslation();
  const [logs, setLogs] = useState([]);
  const [loading, setLoading] = useState(true);
  const [activePage, setActivePage] = useState(1);
  const [totalPages, setTotalPages] = useState(0);
  const [total, setTotal] = useState(0);
  const [selectedLog, setSelectedLog] = useState(null);
  const [showDetail, setShowDetail] = useState(false);

  // 筛选条件
  const [filters, setFilters] = useState({
    external_user_id: '',
    user_group: '',
    start_timestamp: '',
    end_timestamp: '',
  });

  // 排序状态
  const [sortConfig, setSortConfig] = useState({
    key: 'created_at',
    direction: 'desc'
  });

  // 时间选择器状态
  const [timeRange, setTimeRange] = useState('7d'); // 默认7天
  const [customStartDate, setCustomStartDate] = useState('');
  const [customEndDate, setCustomEndDate] = useState('');

  // 时间范围选项
  const timeRangeOptions = [
    { key: '1d', text: '最近1天', value: '1d' },
    { key: '7d', text: '最近7天', value: '7d' },
    { key: '30d', text: '最近30天', value: '30d' },
    { key: '90d', text: '最近90天', value: '90d' },
    { key: 'custom', text: '自定义时间', value: 'custom' },
  ];

  // 更新时间戳
  const updateTimestamps = useCallback(() => {
    const now = new Date();
    let startDate = new Date();

    switch (timeRange) {
      case '1d':
        startDate.setDate(now.getDate() - 1);
        break;
      case '7d':
        startDate.setDate(now.getDate() - 7);
        break;
      case '30d':
        startDate.setDate(now.getDate() - 30);
        break;
      case '90d':
        startDate.setDate(now.getDate() - 90);
        break;
      case 'custom':
        if (customStartDate && customEndDate) {
          const start = new Date(customStartDate);
          const end = new Date(customEndDate);
          setFilters(prev => ({
            ...prev,
            start_timestamp: Math.floor(start.getTime() / 1000).toString(),
            end_timestamp: Math.floor(end.getTime() / 1000).toString(),
          }));
          return;
        }
        return;
      default:
        startDate.setDate(now.getDate() - 7);
    }

    setFilters(prev => ({
      ...prev,
      start_timestamp: Math.floor(startDate.getTime() / 1000).toString(),
      end_timestamp: Math.floor(now.getTime() / 1000).toString(),
    }));
  }, [timeRange, customStartDate, customEndDate]);

  // 时间范围变化时更新时间戳
  useEffect(() => {
    updateTimestamps();
  }, [updateTimestamps]);

  // 加载扩展日志列表
  const loadLogs = useCallback(async (startIdx = 0) => {
    setLoading(true);
    try {
      const params = new URLSearchParams({
        p: Math.floor(startIdx / ITEMS_PER_PAGE),
        page_size: ITEMS_PER_PAGE,
        ...filters,
      });

      const url = isAdmin() ? '/api/log/extended/' : '/api/log/extended/self';
      const res = await API.get(`${url}?${params}`);
      
      if (res.data && res.data.success) {
        setLogs(res.data.data.logs || []);
        setTotal(res.data.data.total || 0);
        setTotalPages(Math.ceil((res.data.data.total || 0) / ITEMS_PER_PAGE));
      } else {
        showError(res.data?.message || '加载失败');
      }
    } catch (error) {
      showError('加载扩展日志失败：' + error.message);
    } finally {
      setLoading(false);
    }
  }, [filters]);

  // 加载日志详情
  const loadLogDetail = async (logId) => {
    try {
      const res = await API.get(`/api/log/extended/${logId}`);
      if (res.data && res.data.success) {
        setSelectedLog(res.data.data);
        setShowDetail(true);
      } else {
        showError('加载日志详情失败：' + (res.data?.message || '未知错误'));
      }
    } catch (error) {
      showError('加载日志详情失败：' + error.message);
    }
  };

  // 初始化加载
  useEffect(() => {
    const timer = setTimeout(() => {
      loadLogs();
    }, 100);

    return () => clearTimeout(timer);
  }, [loadLogs]);

  // 筛选条件变化时重新加载
  useEffect(() => {
    if (activePage === 1) {
      loadLogs(0);
    } else {
      setActivePage(1);
    }
  }, [filters, activePage, loadLogs]);

  // 页码变化时加载
  useEffect(() => {
    loadLogs((activePage - 1) * ITEMS_PER_PAGE);
  }, [activePage, loadLogs]);

  // 处理筛选条件变化
  const handleFilterChange = (key, value) => {
    setFilters(prev => ({
      ...prev,
      [key]: value,
    }));
  };

  // 重置筛选条件
  const resetFilters = () => {
    setFilters({
      external_user_id: '',
      user_group: '',
      start_timestamp: '',
      end_timestamp: '',
    });
    setTimeRange('7d');
    setCustomStartDate('');
    setCustomEndDate('');
  };

  // 处理排序
  const handleSort = (key) => {
    setSortConfig(prev => ({
      key,
      direction: prev.key === key && prev.direction === 'asc' ? 'desc' : 'asc'
    }));
  };

  // 排序数据
  const sortedLogs = useCallback(() => {
    if (!sortConfig.key) return logs;

    return [...logs].sort((a, b) => {
      let aValue = a[sortConfig.key];
      let bValue = b[sortConfig.key];

      // 处理时间字段
      if (sortConfig.key === 'created_at') {
        aValue = new Date(aValue).getTime();
        bValue = new Date(bValue).getTime();
      }

      // 处理字符串字段
      if (typeof aValue === 'string') {
        aValue = aValue.toLowerCase();
        bValue = bValue.toLowerCase();
      }

      if (aValue < bValue) {
        return sortConfig.direction === 'asc' ? -1 : 1;
      }
      if (aValue > bValue) {
        return sortConfig.direction === 'asc' ? 1 : -1;
      }
      return 0;
    });
  }, [logs, sortConfig]);

  // 渲染排序图标
  const renderSortIcon = (key) => {
    if (sortConfig.key !== key) {
      return <Icon name="sort" />;
    }
    return sortConfig.direction === 'asc' ? 
      <Icon name="sort up" /> : 
      <Icon name="sort down" />;
  };

  // 渲染时间戳
  const renderTimestamp = (timestamp, logId) => {
    const formatTimestamp = (ts) => {
      if (!ts) return '-';
      try {
        if (typeof ts === 'string') {
          const date = new Date(ts);
          return date.toLocaleString('zh-CN', {
            year: 'numeric',
            month: '2-digit',
            day: '2-digit',
            hour: '2-digit',
            minute: '2-digit',
            second: '2-digit',
            hour12: false
          });
        }
        return timestamp2string(ts);
      } catch (error) {
        console.error('时间格式化错误:', error);
        return '-';
      }
    };

    return (
      <code
        onClick={async () => {
          if (await copy(logId.toString())) {
            showSuccess(`已复制日志 ID：${logId}`);
          } else {
            showWarning(`日志 ID 复制失败：${logId}`);
          }
        }}
        style={{ cursor: 'pointer' }}
      >
        {formatTimestamp(timestamp)}
      </code>
    );
  };

  // 渲染维度信息
  const renderDimensionInfo = (dimensionInfoStr) => {
    if (!dimensionInfoStr) return '-';
    
    try {
      let dimensionInfo;
      if (typeof dimensionInfoStr === 'string') {
        dimensionInfo = JSON.parse(dimensionInfoStr);
      } else {
        dimensionInfo = dimensionInfoStr;
      }
      
      if (!dimensionInfo || typeof dimensionInfo !== 'object') return '-';
      
      const keys = Object.keys(dimensionInfo);
      if (keys.length === 0) return '-';
      
      const priorityKeys = ['school_name', 'subject_name', 'teacher_name'];
      const displayKeys = [];
      
      priorityKeys.forEach(key => {
        if (dimensionInfo[key] && displayKeys.length < 3) {
          displayKeys.push(key);
        }
      });
      
      keys.forEach(key => {
        if (!priorityKeys.includes(key) && displayKeys.length < 3) {
          displayKeys.push(key);
        }
      });
      
      const parts = displayKeys.map(key => `${key}: ${dimensionInfo[key]}`);
      const result = parts.join(' | ') || '-';
      
      // 如果内容过长，显示省略号
      if (result.length > 50) {
        return (
          <Popup
            content={result}
            trigger={
              <span style={{ cursor: 'pointer' }}>
                {result.substring(0, 50)}...
              </span>
            }
          />
        );
      }
      
      return result;
    } catch (error) {
      console.error('维度信息解析错误:', error);
      return '解析失败';
    }
  };

  // 渲染原始日志信息
  const renderOriginalLogInfo = (originalLog) => {
    if (!originalLog) return '-';
    
    const modelName = originalLog.model_name || '-';
    const quota = originalLog.quota != null ? (typeof t === 'function' ? renderQuota(originalLog.quota, t) : originalLog.quota) : '-';
    const tokens = originalLog.prompt_tokens + originalLog.completion_tokens || 0;
    
    const content = `模型: ${modelName} | 配额: ${quota} | 令牌: ${tokens}`;
    
    // 如果内容过长，显示省略号
    if (content.length > 40) {
      return (
        <Popup
          content={
            <div>
              <div><strong>模型:</strong> {modelName}</div>
              <div><strong>配额:</strong> {quota}</div>
              <div><strong>令牌:</strong> {tokens}</div>
            </div>
          }
          trigger={
            <span style={{ cursor: 'pointer' }}>
              {content.substring(0, 40)}...
            </span>
          }
        />
      );
    }
    
    return content;
  };

  // 渲染筛选器
  const renderFilters = () => {
    return (
      <Segment>
        <Header size="small">筛选条件</Header>
        <Form>
          <Form.Group widths="equal">
            <Form.Field>
              <label>时间范围</label>
              <Dropdown
                fluid
                selection
                options={timeRangeOptions}
                value={timeRange}
                onChange={(e, { value }) => setTimeRange(value)}
              />
            </Form.Field>
            <Form.Field>
              <label>外部用户ID</label>
              <Input
                placeholder="输入教师ID"
                value={filters.external_user_id}
                onChange={(e) => handleFilterChange('external_user_id', e.target.value)}
              />
            </Form.Field>
            <Form.Field>
              <label>用户组</label>
              <Input
                placeholder="输入用户组"
                value={filters.user_group}
                onChange={(e) => handleFilterChange('user_group', e.target.value)}
              />
            </Form.Field>
            <Form.Field>
              <label>&nbsp;</label>
              <Button onClick={resetFilters}>重置</Button>
            </Form.Field>
          </Form.Group>
          
          {timeRange === 'custom' && (
            <Form.Group widths="equal">
              <Form.Field>
                <label>开始日期</label>
                <Input
                  type="date"
                  value={customStartDate}
                  onChange={(e) => setCustomStartDate(e.target.value)}
                />
              </Form.Field>
              <Form.Field>
                <label>结束日期</label>
                <Input
                  type="date"
                  value={customEndDate}
                  onChange={(e) => setCustomEndDate(e.target.value)}
                />
              </Form.Field>
            </Form.Group>
          )}
        </Form>
      </Segment>
    );
  };

  // 渲染日志详情模态框
  const renderDetailModal = () => {
    if (!selectedLog) return null;

    return (
      <Modal open={showDetail} onClose={() => setShowDetail(false)} size="large">
        <Modal.Header>扩展日志详情</Modal.Header>
        <Modal.Content>
          <Grid columns={2} divided>
            <Grid.Column>
              <Header size="small">扩展日志信息</Header>
              <Table basic="very" celled>
                <Table.Body>
                  <Table.Row>
                    <Table.Cell><strong>ID</strong></Table.Cell>
                    <Table.Cell>{selectedLog.id}</Table.Cell>
                  </Table.Row>
                  <Table.Row>
                    <Table.Cell><strong>原始日志ID</strong></Table.Cell>
                    <Table.Cell>{selectedLog.log_id}</Table.Cell>
                  </Table.Row>
                  <Table.Row>
                    <Table.Cell><strong>外部用户ID</strong></Table.Cell>
                    <Table.Cell>{selectedLog.external_user_id}</Table.Cell>
                  </Table.Row>
                  <Table.Row>
                    <Table.Cell><strong>用户组</strong></Table.Cell>
                    <Table.Cell>{renderColorLabel(selectedLog.user_group)}</Table.Cell>
                  </Table.Row>
                  <Table.Row>
                    <Table.Cell><strong>创建时间</strong></Table.Cell>
                    <Table.Cell>{selectedLog.created_at ? new Date(selectedLog.created_at).toLocaleString('zh-CN') : '-'}</Table.Cell>
                  </Table.Row>
                </Table.Body>
              </Table>

              {selectedLog.dimension_info && typeof selectedLog.dimension_info === 'object' && (
                <>
                  <Header size="small">维度信息</Header>
                  <Table basic="very" celled>
                    <Table.Body>
                      {Object.entries(selectedLog.dimension_info).map(([key, value]) => (
                        <Table.Row key={key}>
                          <Table.Cell><strong>{key}</strong></Table.Cell>
                          <Table.Cell>{value}</Table.Cell>
                        </Table.Row>
                      ))}
                    </Table.Body>
                  </Table>
                </>
              )}
            </Grid.Column>

            <Grid.Column>
              <Header size="small">原始日志信息</Header>
              {selectedLog.original_log ? (
                <Table basic="very" celled>
                  <Table.Body>
                    <Table.Row>
                      <Table.Cell><strong>用户ID</strong></Table.Cell>
                      <Table.Cell>{selectedLog.original_log.user_id}</Table.Cell>
                    </Table.Row>
                    <Table.Row>
                      <Table.Cell><strong>用户名</strong></Table.Cell>
                      <Table.Cell>{selectedLog.original_log.username}</Table.Cell>
                    </Table.Row>
                    <Table.Row>
                      <Table.Cell><strong>模型名称</strong></Table.Cell>
                      <Table.Cell>{selectedLog.original_log.model_name}</Table.Cell>
                    </Table.Row>
                    <Table.Row>
                      <Table.Cell><strong>令牌名称</strong></Table.Cell>
                      <Table.Cell>{selectedLog.original_log.token_name}</Table.Cell>
                    </Table.Row>
                    <Table.Row>
                      <Table.Cell><strong>配额</strong></Table.Cell>
                      <Table.Cell>
                        {selectedLog.original_log.quota != null ? (
                          typeof t === 'function' ? renderQuota(selectedLog.original_log.quota, t) : selectedLog.original_log.quota
                        ) : '-'}
                      </Table.Cell>
                    </Table.Row>
                    <Table.Row>
                      <Table.Cell><strong>提示令牌</strong></Table.Cell>
                      <Table.Cell>{selectedLog.original_log.prompt_tokens}</Table.Cell>
                    </Table.Row>
                    <Table.Row>
                      <Table.Cell><strong>完成令牌</strong></Table.Cell>
                      <Table.Cell>{selectedLog.original_log.completion_tokens}</Table.Cell>
                    </Table.Row>
                    <Table.Row>
                      <Table.Cell><strong>渠道ID</strong></Table.Cell>
                      <Table.Cell>{selectedLog.original_log.channel}</Table.Cell>
                    </Table.Row>
                    <Table.Row>
                      <Table.Cell><strong>请求ID</strong></Table.Cell>
                      <Table.Cell>
                        <code
                          onClick={async () => {
                            if (await copy(selectedLog.original_log.request_id)) {
                              showSuccess('已复制请求ID');
                            }
                          }}
                          style={{ cursor: 'pointer' }}
                        >
                          {selectedLog.original_log.request_id}
                        </code>
                      </Table.Cell>
                    </Table.Row>
                  </Table.Body>
                </Table>
              ) : (
                <Message warning>
                  <Message.Header>原始日志不存在</Message.Header>
                  <p>该扩展日志对应的原始日志可能已被删除。</p>
                </Message>
              )}
            </Grid.Column>
          </Grid>
        </Modal.Content>
        <Modal.Actions>
          <Button onClick={() => setShowDetail(false)}>关闭</Button>
        </Modal.Actions>
      </Modal>
    );
  };

  return (
    <div className="dashboard-container">
      {/* 筛选器 */}
      {renderFilters()}

      {/* 日志列表 */}
      <Card fluid>
        <Card.Content>
          <Card.Header>
            <Icon name="list" />
            扩展日志列表
          </Card.Header>
          
          <Table celled sortable>
            <Table.Header>
              <Table.Row>
                <Table.HeaderCell 
                  sorted={sortConfig.key === 'created_at' ? sortConfig.direction : null}
                  onClick={() => handleSort('created_at')}
                  style={{ cursor: 'pointer', width: '15%' }}
                >
                  时间 {renderSortIcon('created_at')}
                </Table.HeaderCell>
                <Table.HeaderCell 
                  sorted={sortConfig.key === 'external_user_id' ? sortConfig.direction : null}
                  onClick={() => handleSort('external_user_id')}
                  style={{ cursor: 'pointer', width: '15%' }}
                >
                  外部用户ID {renderSortIcon('external_user_id')}
                </Table.HeaderCell>
                <Table.HeaderCell 
                  sorted={sortConfig.key === 'user_group' ? sortConfig.direction : null}
                  onClick={() => handleSort('user_group')}
                  style={{ cursor: 'pointer', width: '15%' }}
                >
                  用户组 {renderSortIcon('user_group')}
                </Table.HeaderCell>
                <Table.HeaderCell style={{ width: '25%' }}>维度信息</Table.HeaderCell>
                <Table.HeaderCell style={{ width: '20%' }}>原始日志</Table.HeaderCell>
                <Table.HeaderCell style={{ width: '10%' }}>操作</Table.HeaderCell>
              </Table.Row>
            </Table.Header>

            <Table.Body>
              {sortedLogs().length === 0 ? (
                <Table.Row>
                  <Table.Cell colSpan="6" textAlign="center" style={{ padding: '2rem' }}>
                    {loading ? (
                      <div>
                        <Icon name="spinner" loading />
                        <div style={{ marginTop: '0.5rem' }}>加载中...</div>
                      </div>
                    ) : (
                      <div>
                        <Icon name="inbox" size="large" style={{ color: '#ccc' }} />
                        <div style={{ marginTop: '0.5rem', color: '#666' }}>暂无数据</div>
                      </div>
                    )}
                  </Table.Cell>
                </Table.Row>
              ) : (
                sortedLogs().map((log) => (
                  <Table.Row key={log.id}>
                    <Table.Cell style={{ whiteSpace: 'nowrap' }}>
                      {renderTimestamp(log.created_at, log.id)}
                    </Table.Cell>
                    <Table.Cell style={{ whiteSpace: 'nowrap' }}>
                      <code>{log.external_user_id}</code>
                    </Table.Cell>
                    <Table.Cell style={{ whiteSpace: 'nowrap' }}>
                      {renderColorLabel(log.user_group)}
                    </Table.Cell>
                    <Table.Cell style={{ whiteSpace: 'normal', wordBreak: 'break-all' }}>
                      {renderDimensionInfo(log.dimension_info)}
                    </Table.Cell>
                    <Table.Cell style={{ whiteSpace: 'normal', wordBreak: 'break-all' }}>
                      {renderOriginalLogInfo(log.original_log)}
                    </Table.Cell>
                    <Table.Cell style={{ whiteSpace: 'nowrap', textAlign: 'center' }}>
                      <Popup
                        content="查看详情"
                        trigger={
                          <Button
                            size="small"
                            icon="eye"
                            onClick={() => loadLogDetail(log.log_id)}
                          />
                        }
                      />
                    </Table.Cell>
                  </Table.Row>
                ))
              )}
            </Table.Body>
          </Table>

          {/* 分页器 */}
          {totalPages > 1 && (
            <div style={{ display: 'flex', justifyContent: 'center', marginTop: '1rem' }}>
              <Pagination
                activePage={activePage}
                onPageChange={(e, { activePage }) => setActivePage(activePage)}
                totalPages={totalPages}
                siblingRange={1}
                showFirstAndLastNav
                showPreviousAndNextNav
              />
            </div>
          )}

          {/* 总数显示 */}
          <div style={{ marginTop: '1rem', textAlign: 'center', color: '#666' }}>
            共 {total} 条记录
          </div>
        </Card.Content>
      </Card>

      {/* 详情模态框 */}
      {renderDetailModal()}
    </div>
  );
};

export default ExtendedLogs;