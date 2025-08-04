import React, { useEffect, useState, useCallback } from 'react';
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
  Statistic,
  Divider,
  Icon,
  Message,
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
  const [logs, setLogs] = useState([]);
  const [loading, setLoading] = useState(true);
  const [activePage, setActivePage] = useState(1);
  const [totalPages, setTotalPages] = useState(0);
  const [total, setTotal] = useState(0);
  const [statistics, setStatistics] = useState(null);
  const [showStatistics, setShowStatistics] = useState(false);
  const [selectedLog, setSelectedLog] = useState(null);
  const [showDetail, setShowDetail] = useState(false);

  // 筛选条件
  const [filters, setFilters] = useState({
    external_user_id: '',
    user_group: '',
    school_id: '',
    subject_id: '',
    start_timestamp: '',
    end_timestamp: '',
  });

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
      
      if (res.success) {
        setLogs(res.data.logs || []);
        setTotal(res.data.total || 0);
        setTotalPages(Math.ceil((res.data.total || 0) / ITEMS_PER_PAGE));
      } else {
        showError(res.message);
      }
    } catch (error) {
      showError('加载扩展日志失败：' + error.message);
    } finally {
      setLoading(false);
    }
  }, [filters]);

  // 加载统计信息
  const loadStatistics = useCallback(async () => {
    if (!isAdmin()) return;
    
    try {
      const params = new URLSearchParams();
      if (filters.start_timestamp) {
        params.append('start_timestamp', filters.start_timestamp);
      }
      if (filters.end_timestamp) {
        params.append('end_timestamp', filters.end_timestamp);
      }

      const res = await API.get(`/api/log/extended/statistics?${params}`);
      if (res.success) {
        setStatistics(res.data);
      } else {
        showError('加载统计信息失败：' + res.message);
      }
    } catch (error) {
      showError('加载统计信息失败：' + error.message);
    }
  }, [filters.start_timestamp, filters.end_timestamp]);

  // 加载日志详情
  const loadLogDetail = async (logId) => {
    try {
      const res = await API.get(`/api/log/extended/${logId}`);
      if (res.success) {
        setSelectedLog(res.data);
        setShowDetail(true);
      } else {
        showError('加载日志详情失败：' + res.message);
      }
    } catch (error) {
      showError('加载日志详情失败：' + error.message);
    }
  };

  // 初始化加载
  useEffect(() => {
    loadLogs();
    if (isAdmin()) {
      loadStatistics();
    }
  }, [loadLogs, loadStatistics]);

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
      school_id: '',
      subject_id: '',
      start_timestamp: '',
      end_timestamp: '',
    });
  };

  // 渲染时间戳
  const renderTimestamp = (timestamp, logId) => {
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
        {timestamp2string(timestamp)}
      </code>
    );
  };

  // 渲染维度信息
  const renderDimensionInfo = (dimensionInfo) => {
    if (!dimensionInfo) return '-';
    
    const parts = [];
    if (dimensionInfo.school_name) {
      parts.push(dimensionInfo.school_name);
    }
    if (dimensionInfo.subject_name) {
      parts.push(dimensionInfo.subject_name);
    }
    if (dimensionInfo.teacher_name) {
      parts.push(dimensionInfo.teacher_name);
    }
    
    return parts.join(' / ') || '-';
  };

  // 渲染原始日志信息
  const renderOriginalLogInfo = (originalLog) => {
    if (!originalLog) return '-';
    
    return (
      <div>
        <div><strong>模型:</strong> {originalLog.model_name || '-'}</div>
        <div><strong>配额:</strong> {renderQuota(originalLog.quota)}</div>
        <div><strong>令牌:</strong> {originalLog.prompt_tokens + originalLog.completion_tokens || 0}</div>
      </div>
    );
  };

  // 渲染统计卡片
  const renderStatisticsCard = () => {
    if (!statistics) return null;

    return (
      <Card fluid>
        <Card.Content>
          <Card.Header>
            <Icon name="chart bar" />
            扩展日志统计
            <Button
              floated="right"
              size="small"
              onClick={() => setShowStatistics(!showStatistics)}
            >
              {showStatistics ? '隐藏' : '显示'}
            </Button>
          </Card.Header>
          {showStatistics && (
            <Card.Description>
              <Grid columns={4} divided>
                <Grid.Column>
                  <Statistic size="small">
                    <Statistic.Value>{statistics.total_count || 0}</Statistic.Value>
                    <Statistic.Label>总记录数</Statistic.Label>
                  </Statistic>
                </Grid.Column>
                <Grid.Column>
                  <Statistic size="small">
                    <Statistic.Value>{statistics.user_group_stats?.length || 0}</Statistic.Value>
                    <Statistic.Label>用户组数</Statistic.Label>
                  </Statistic>
                </Grid.Column>
                <Grid.Column>
                  <Statistic size="small">
                    <Statistic.Value>{statistics.school_stats?.length || 0}</Statistic.Value>
                    <Statistic.Label>学校数</Statistic.Label>
                  </Statistic>
                </Grid.Column>
                <Grid.Column>
                  <Statistic size="small">
                    <Statistic.Value>{statistics.subject_stats?.length || 0}</Statistic.Value>
                    <Statistic.Label>学科组数</Statistic.Label>
                  </Statistic>
                </Grid.Column>
              </Grid>

              <Divider />

              <Grid columns={3}>
                <Grid.Column>
                  <Header size="small">用户组Top5</Header>
                  {statistics.user_group_stats?.slice(0, 5).map((item, index) => (
                    <div key={index}>
                      <Label size="small">
                        {item.user_group}: {item.count}
                      </Label>
                    </div>
                  ))}
                </Grid.Column>
                <Grid.Column>
                  <Header size="small">学校Top5</Header>
                  {statistics.school_stats?.slice(0, 5).map((item, index) => (
                    <div key={index}>
                      <Label size="small">
                        {item.school_name}: {item.count}
                      </Label>
                    </div>
                  ))}
                </Grid.Column>
                <Grid.Column>
                  <Header size="small">学科组Top5</Header>
                  {statistics.subject_stats?.slice(0, 5).map((item, index) => (
                    <div key={index}>
                      <Label size="small">
                        {item.subject_name}: {item.count}
                      </Label>
                    </div>
                  ))}
                </Grid.Column>
              </Grid>
            </Card.Description>
          )}
        </Card.Content>
      </Card>
    );
  };

  // 渲染筛选器
  const renderFilters = () => {
    return (
      <Segment>
        <Header size="small">筛选条件</Header>
        <Form>
          <Form.Group widths="equal">
            <Form.Input
              label="外部用户ID"
              placeholder="输入教师ID"
              value={filters.external_user_id}
              onChange={(e) => handleFilterChange('external_user_id', e.target.value)}
            />
            <Form.Input
              label="用户组"
              placeholder="输入用户组"
              value={filters.user_group}
              onChange={(e) => handleFilterChange('user_group', e.target.value)}
            />
            <Form.Input
              label="学校ID"
              placeholder="输入学校ID"
              value={filters.school_id}
              onChange={(e) => handleFilterChange('school_id', e.target.value)}
            />
            <Form.Input
              label="学科组ID"
              placeholder="输入学科组ID"
              value={filters.subject_id}
              onChange={(e) => handleFilterChange('subject_id', e.target.value)}
            />
          </Form.Group>
          <Form.Group>
            <Form.Input
              label="开始时间戳"
              placeholder="Unix时间戳"
              value={filters.start_timestamp}
              onChange={(e) => handleFilterChange('start_timestamp', e.target.value)}
            />
            <Form.Input
              label="结束时间戳"
              placeholder="Unix时间戳"
              value={filters.end_timestamp}
              onChange={(e) => handleFilterChange('end_timestamp', e.target.value)}
            />
            <Form.Field>
              <label>&nbsp;</label>
              <Button onClick={resetFilters}>重置</Button>
            </Form.Field>
          </Form.Group>
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
                    <Table.Cell>{timestamp2string(selectedLog.created_at)}</Table.Cell>
                  </Table.Row>
                </Table.Body>
              </Table>

              {selectedLog.dimension_info && (
                <>
                  <Header size="small">维度信息</Header>
                  <Table basic="very" celled>
                    <Table.Body>
                      {selectedLog.dimension_info.school_name && (
                        <Table.Row>
                          <Table.Cell><strong>学校</strong></Table.Cell>
                          <Table.Cell>{selectedLog.dimension_info.school_name}</Table.Cell>
                        </Table.Row>
                      )}
                      {selectedLog.dimension_info.subject_name && (
                        <Table.Row>
                          <Table.Cell><strong>学科组</strong></Table.Cell>
                          <Table.Cell>{selectedLog.dimension_info.subject_name}</Table.Cell>
                        </Table.Row>
                      )}
                      {selectedLog.dimension_info.teacher_name && (
                        <Table.Row>
                          <Table.Cell><strong>教师姓名</strong></Table.Cell>
                          <Table.Cell>{selectedLog.dimension_info.teacher_name}</Table.Cell>
                        </Table.Row>
                      )}
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
                      <Table.Cell>{renderQuota(selectedLog.original_log.quota)}</Table.Cell>
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
                      <Table.Cell>{selectedLog.original_log.channel_id}</Table.Cell>
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
      {/* 统计信息卡片 */}
      {isAdmin() && renderStatisticsCard()}

      {/* 筛选器 */}
      {renderFilters()}

      {/* 日志列表 */}
      <Card fluid>
        <Card.Content>
          <Card.Header>
            <Icon name="list" />
            扩展日志列表
            <Button
              floated="right"
              size="small"
              onClick={() => loadLogs((activePage - 1) * ITEMS_PER_PAGE)}
              loading={loading}
            >
              刷新
            </Button>
          </Card.Header>
          
          <Table celled>
            <Table.Header>
              <Table.Row>
                <Table.HeaderCell>时间</Table.HeaderCell>
                <Table.HeaderCell>外部用户ID</Table.HeaderCell>
                <Table.HeaderCell>用户组</Table.HeaderCell>
                <Table.HeaderCell>维度信息</Table.HeaderCell>
                <Table.HeaderCell>原始日志</Table.HeaderCell>
                <Table.HeaderCell>操作</Table.HeaderCell>
              </Table.Row>
            </Table.Header>

            <Table.Body>
              {logs.length === 0 ? (
                <Table.Row>
                  <Table.Cell colSpan="6" textAlign="center">
                    {loading ? '加载中...' : '暂无数据'}
                  </Table.Cell>
                </Table.Row>
              ) : (
                logs.map((log) => (
                  <Table.Row key={log.id}>
                    <Table.Cell>
                      {renderTimestamp(log.created_at, log.id)}
                    </Table.Cell>
                    <Table.Cell>
                      <code>{log.external_user_id}</code>
                    </Table.Cell>
                    <Table.Cell>
                      {renderColorLabel(log.user_group)}
                    </Table.Cell>
                    <Table.Cell>
                      {renderDimensionInfo(log.dimension_info)}
                    </Table.Cell>
                    <Table.Cell>
                      {renderOriginalLogInfo(log.original_log)}
                    </Table.Cell>
                    <Table.Cell>
                      <Button
                        size="small"
                        onClick={() => loadLogDetail(log.log_id)}
                      >
                        详情
                      </Button>
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