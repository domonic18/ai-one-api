import React, { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import {
  Button,
  Container,
  Form,
  Header,
  Icon,
  Segment,
  Message,
  Label,
  Grid,
  Statistic,
  Table,
  Modal,
  Confirm,
  Divider,
  Card,
  Progress,
} from 'semantic-ui-react';
import { API, showError, showSuccess } from '../../helpers';
import './CoursewareIntegration.css';

const CoursewareIntegration = () => {
  const { t } = useTranslation();
  const [config, setConfig] = useState({
    enabled: false,
    base_url: '',
    api_key: '',
    timeout: 5,
    cache_ttl: 10,
    default_group: 'default',
    preload_batch_size: 100,
    refresh_interval: 60,
  });
  const [status, setStatus] = useState({
    isConnected: false,
    lastSyncTime: null,
    totalUsers: 0,
    cachedUsers: 0,
    syncProgress: 0,
    isSyncing: false,
  });
  const [cacheData, setCacheData] = useState([]);
  const [loading, setLoading] = useState(true);
  const [modalOpen, setModalOpen] = useState(false);
  const [confirmOpen, setConfirmOpen] = useState(false);
  const [editingCache, setEditingCache] = useState(null);

  useEffect(() => {
    loadConfig();
    loadStatus();
    loadCacheData();
  }, []);

  const loadConfig = async () => {
    try {
      const response = await API.get('/api/courseware/config');
      if (response.data.success) {
        setConfig(response.data.data || {});
      }
    } catch (error) {
      console.error('加载配置失败:', error);
    }
  };

  const loadStatus = async () => {
    try {
      const response = await API.get('/api/courseware/status');
      if (response.data.success) {
        setStatus(response.data.data || {});
      }
    } catch (error) {
      console.error('加载状态失败:', error);
    }
  };

  const loadCacheData = async () => {
    try {
      setLoading(true);
      const response = await API.get('/api/courseware/cache');
      if (response.data.success) {
        setCacheData(response.data.data || []);
      }
    } catch (error) {
      console.error('加载缓存数据失败:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleConfigSubmit = async () => {
    try {
      const response = await API.post('/api/courseware/config', config);
      if (response.data.success) {
        showSuccess('配置保存成功');
        setModalOpen(false);
        loadConfig();
        loadStatus();
      } else {
        showError(response.data.message || '保存失败');
      }
    } catch (error) {
      showError('保存失败: ' + error.message);
    }
  };

  const handleTestConnection = async () => {
    try {
      const response = await API.post('/api/courseware/test');
      if (response.data.success) {
        showSuccess('连接测试成功');
        loadStatus();
      } else {
        showError(response.data.message || '连接测试失败');
      }
    } catch (error) {
      showError('连接测试失败: ' + error.message);
    }
  };

  const handleSyncUsers = async () => {
    try {
      setStatus(prev => ({ ...prev, isSyncing: true }));
      const response = await API.post('/api/courseware/sync');
      if (response.data.success) {
        showSuccess('用户同步成功');
        loadStatus();
        loadCacheData();
      } else {
        showError(response.data.message || '同步失败');
      }
    } catch (error) {
      showError('同步失败: ' + error.message);
    } finally {
      setStatus(prev => ({ ...prev, isSyncing: false }));
    }
  };

  const handleClearCache = async () => {
    try {
      const response = await API.delete('/api/courseware/cache');
      if (response.data.success) {
        showSuccess('缓存清理成功');
        loadStatus();
        loadCacheData();
      } else {
        showError(response.data.message || '清理失败');
      }
    } catch (error) {
      showError('清理失败: ' + error.message);
    }
  };

  const handleDeleteCache = async () => {
    try {
      const response = await API.delete(`/api/courseware/cache/${editingCache.teacher_id}`);
      if (response.data.success) {
        showSuccess('缓存删除成功');
        setConfirmOpen(false);
        setEditingCache(null);
        loadCacheData();
      } else {
        showError(response.data.message || '删除失败');
      }
    } catch (error) {
      showError('删除失败: ' + error.message);
    }
  };

  const renderConnectionStatus = () => {
    if (status.isConnected) {
      return <Label color="green">已连接</Label>;
    }
    return <Label color="red">未连接</Label>;
  };

  const renderSyncStatus = () => {
    if (status.isSyncing) {
      return (
        <div>
          <Label color="blue">同步中</Label>
          <Progress percent={status.syncProgress} indicating />
        </div>
      );
    }
    return <Label color="grey">空闲</Label>;
  };

  return (
    <Container>
      <Header as="h2" icon textAlign="center">
        <Icon name="cloud" circular />
        <Header.Content>{t('courseware.title')}</Header.Content>
      </Header>

      {/* 状态概览 */}
      <Grid columns={4} stackable style={{ marginBottom: '20px' }}>
        <Grid.Column>
          <Statistic>
            <Statistic.Value>{renderConnectionStatus()}</Statistic.Value>
            <Statistic.Label>连接状态</Statistic.Label>
          </Statistic>
        </Grid.Column>
        <Grid.Column>
          <Statistic>
            <Statistic.Value>{status.totalUsers}</Statistic.Value>
            <Statistic.Label>总用户数</Statistic.Label>
          </Statistic>
        </Grid.Column>
        <Grid.Column>
          <Statistic>
            <Statistic.Value>{status.cachedUsers}</Statistic.Value>
            <Statistic.Label>缓存用户数</Statistic.Label>
          </Statistic>
        </Grid.Column>
        <Grid.Column>
          <Statistic>
            <Statistic.Value>{renderSyncStatus()}</Statistic.Value>
            <Statistic.Label>同步状态</Statistic.Label>
          </Statistic>
        </Grid.Column>
      </Grid>

      {/* 配置卡片 */}
      <Card fluid>
        <Card.Content>
          <Card.Header>
            <Icon name="settings" />
            集成配置
          </Card.Header>
          <Card.Description>
            <Grid columns={2} stackable>
              <Grid.Column>
                <p><strong>启用状态:</strong> {config.enabled ? '已启用' : '已禁用'}</p>
                <p><strong>API地址:</strong> {config.base_url || '未配置'}</p>
                <p><strong>超时时间:</strong> {config.timeout}秒</p>
              </Grid.Column>
              <Grid.Column>
                <p><strong>缓存时间:</strong> {config.cache_ttl}分钟</p>
                <p><strong>默认分组:</strong> {config.default_group}</p>
                <p><strong>刷新间隔:</strong> {config.refresh_interval}分钟</p>
              </Grid.Column>
            </Grid>
          </Card.Description>
        </Card.Content>
        <Card.Content extra>
          <Button.Group>
            <Button primary onClick={() => setModalOpen(true)}>
              <Icon name="edit" />
              编辑配置
            </Button>
            <Button onClick={handleTestConnection}>
              <Icon name="wifi" />
              测试连接
            </Button>
          </Button.Group>
        </Card.Content>
      </Card>

      <Divider />

      {/* 操作按钮 */}
      <Segment>
        <Button.Group>
          <Button primary onClick={handleSyncUsers} loading={status.isSyncing}>
            <Icon name="sync" />
            同步用户
          </Button>
          <Button onClick={handleClearCache}>
            <Icon name="trash" />
            清理缓存
          </Button>
          <Button onClick={loadCacheData}>
            <Icon name="refresh" />
            刷新数据
          </Button>
        </Button.Group>
      </Segment>

      {/* 缓存数据表格 */}
      <Segment>
        <Header as="h3">
          <Icon name="database" />
          缓存数据
        </Header>
        
        <Table celled>
          <Table.Header>
            <Table.Row>
              <Table.HeaderCell>用户ID</Table.HeaderCell>
              <Table.HeaderCell>用户组</Table.HeaderCell>
              <Table.HeaderCell>偏好模型</Table.HeaderCell>
              <Table.HeaderCell>学校</Table.HeaderCell>
              <Table.HeaderCell>学科组</Table.HeaderCell>
              <Table.HeaderCell>更新时间</Table.HeaderCell>
              <Table.HeaderCell>操作</Table.HeaderCell>
            </Table.Row>
          </Table.Header>

          <Table.Body>
            {cacheData.map((item) => (
              <Table.Row key={item.teacher_id}>
                <Table.Cell>{item.teacher_id}</Table.Cell>
                <Table.Cell>{item.group_name}</Table.Cell>
                <Table.Cell>{item.preferred_model || '-'}</Table.Cell>
                <Table.Cell>{item.school_name || '-'}</Table.Cell>
                <Table.Cell>{item.subject_name || '-'}</Table.Cell>
                <Table.Cell>{new Date(item.updated_at * 1000).toLocaleString()}</Table.Cell>
                <Table.Cell>
                  <Button.Group size="mini">
                    <Button
                      icon
                      onClick={() => {
                        setEditingCache(item);
                        setConfirmOpen(true);
                      }}
                      title="删除缓存"
                    >
                      <Icon name="trash" />
                    </Button>
                  </Button.Group>
                </Table.Cell>
              </Table.Row>
            ))}
          </Table.Body>
        </Table>

        {cacheData.length === 0 && !loading && (
          <Message info>
            <Message.Header>暂无缓存数据</Message.Header>
            <p>点击"同步用户"按钮开始同步数据</p>
          </Message>
        )}
      </Segment>

      {/* 配置编辑模态框 */}
      <Modal open={modalOpen} onClose={() => setModalOpen(false)} size="large">
        <Modal.Header>编辑集成配置</Modal.Header>
        <Modal.Content>
          <Form>
            <Form.Field>
              <label>启用集成</label>
              <Form.Checkbox
                label="启用课件平台集成"
                checked={config.enabled}
                onChange={(e, { checked }) => setConfig({ ...config, enabled: checked })}
              />
            </Form.Field>
            <Form.Field>
              <label>API地址</label>
              <Form.Input
                placeholder="https://courseware.example.com/api/v1"
                value={config.base_url}
                onChange={(e, { value }) => setConfig({ ...config, base_url: value })}
              />
            </Form.Field>
            <Form.Field>
              <label>API密钥</label>
              <Form.Input
                type="password"
                placeholder="请输入API密钥"
                value={config.api_key}
                onChange={(e, { value }) => setConfig({ ...config, api_key: value })}
              />
            </Form.Field>
            <Form.Field>
              <label>超时时间（秒）</label>
              <Form.Input
                type="number"
                value={config.timeout}
                onChange={(e, { value }) => setConfig({ ...config, timeout: parseInt(value) || 5 })}
              />
            </Form.Field>
            <Form.Field>
              <label>缓存时间（分钟）</label>
              <Form.Input
                type="number"
                value={config.cache_ttl}
                onChange={(e, { value }) => setConfig({ ...config, cache_ttl: parseInt(value) || 10 })}
              />
            </Form.Field>
            <Form.Field>
              <label>默认分组</label>
              <Form.Input
                placeholder="default"
                value={config.default_group}
                onChange={(e, { value }) => setConfig({ ...config, default_group: value })}
              />
            </Form.Field>
            <Form.Field>
              <label>预加载批次大小</label>
              <Form.Input
                type="number"
                value={config.preload_batch_size}
                onChange={(e, { value }) => setConfig({ ...config, preload_batch_size: parseInt(value) || 100 })}
              />
            </Form.Field>
            <Form.Field>
              <label>刷新间隔（分钟）</label>
              <Form.Input
                type="number"
                value={config.refresh_interval}
                onChange={(e, { value }) => setConfig({ ...config, refresh_interval: parseInt(value) || 60 })}
              />
            </Form.Field>
          </Form>
        </Modal.Content>
        <Modal.Actions>
          <Button onClick={() => setModalOpen(false)}>
            取消
          </Button>
          <Button primary onClick={handleConfigSubmit}>
            保存
          </Button>
        </Modal.Actions>
      </Modal>

      {/* 删除确认框 */}
      <Confirm
        open={confirmOpen}
        header="删除缓存"
        content={`确定要删除用户 "${editingCache?.teacher_id}" 的缓存数据吗？`}
        onCancel={() => setConfirmOpen(false)}
        onConfirm={handleDeleteCache}
        cancelButton="取消"
        confirmButton="删除"
      />
    </Container>
  );
};

export default CoursewareIntegration; 