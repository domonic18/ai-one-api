import React, { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import {
  Button,
  Container,
  Form,
  Header,
  Icon,
  Modal,
  Segment,
  Table,
  Message,
  Confirm,
  Label,
  Grid,
  Statistic,
} from 'semantic-ui-react';
import { API, showError, showSuccess } from '../../helpers';
import './GroupManagement.css';

const GroupManagement = () => {
  const { t } = useTranslation();
  const [groups, setGroups] = useState([]);
  const [loading, setLoading] = useState(true);
  const [modalOpen, setModalOpen] = useState(false);
  const [confirmOpen, setConfirmOpen] = useState(false);
  const [editingGroup, setEditingGroup] = useState(null);
  const [formData, setFormData] = useState({
    name: '',
    description: '',
    total_quota: 0,
    user_count: 0,
  });
  const [stats, setStats] = useState({
    totalGroups: 0,
    totalUsers: 0,
    totalQuota: 0,
  });

  useEffect(() => {
    loadGroups();
  }, []);

  const loadGroups = async () => {
    try {
      setLoading(true);
      // 使用新的详细API端点
      const response = await API.get('/api/group/detail');
      if (response.data.success) {
        const groupData = response.data.data || [];
        // 使用用户组名称作为ID
        const groupsWithId = groupData.map((group, index) => ({
          ...group,
          id: group.name, // 使用用户组名称作为ID
          index: index + 1, // 保留索引用于显示
        }));
        setGroups(groupsWithId);
        calculateStats(groupsWithId);
      } else {
        showError(response.data.message || '加载用户组失败');
      }
    } catch (error) {
      showError('加载用户组失败: ' + error.message);
    } finally {
      setLoading(false);
    }
  };

  const calculateStats = (groupData) => {
    const stats = {
      totalGroups: groupData.length,
      totalUsers: groupData.reduce((sum, group) => sum + (group.user_count || 0), 0),
      totalQuota: groupData.reduce((sum, group) => sum + (group.total_quota || 0), 0),
    };
    setStats(stats);
  };

  const handleSubmit = async () => {
    try {
      const url = editingGroup ? `/api/group/${editingGroup.id}` : '/api/group';
      const method = editingGroup ? 'put' : 'post';
      
      const response = await API[method](url, formData);
      if (response.data.success) {
        showSuccess(editingGroup ? '用户组更新成功' : '用户组创建成功');
        setModalOpen(false);
        resetForm();
        loadGroups();
      } else {
        showError(response.data.message || '操作失败');
      }
    } catch (error) {
      // 优先显示后端返回的错误信息
      if (error.response && error.response.data && error.response.data.message) {
        showError(error.response.data.message);
      } else if (error.response && error.response.status) {
        showError(`操作失败 (${error.response.status}): ${error.response.statusText || '服务器错误'}`);
      } else {
        showError('操作失败: ' + error.message);
      }
    }
  };

  const handleDelete = async () => {
    try {
      const response = await API.delete(`/api/group/${editingGroup.id}`);
      if (response.data.success) {
        showSuccess('用户组删除成功');
        setConfirmOpen(false);
        setEditingGroup(null);
        loadGroups();
      } else {
        showError(response.data.message || '删除失败');
      }
    } catch (error) {
      // 优先显示后端返回的错误信息
      if (error.response && error.response.data && error.response.data.message) {
        showError(error.response.data.message);
      } else if (error.response && error.response.status) {
        showError(`删除失败 (${error.response.status}): ${error.response.statusText || '服务器错误'}`);
      } else {
        showError('删除失败: ' + error.message);
      }
    }
  };

  const openEditModal = (group) => {
    setEditingGroup(group);
    setFormData({
      name: group.name,
      description: group.description || '',
      total_quota: group.total_quota || 0,
      user_count: group.user_count || 0,
    });
    setModalOpen(true);
  };

  const openCreateModal = () => {
    setEditingGroup(null);
    resetForm();
    setModalOpen(true);
  };

  const resetForm = () => {
    setFormData({
      name: '',
      description: '',
      total_quota: 0,
      user_count: 0,
    });
  };

  const renderGroupStatus = (group) => {
    if (group.user_count > 0) {
      return <Label color="green">活跃</Label>;
    }
    return <Label color="grey">空闲</Label>;
  };

  return (
    <Container>
      <Header as="h2" icon textAlign="center">
        <Icon name="users" circular />
        <Header.Content>{t('group.title')}</Header.Content>
      </Header>

      {/* 统计信息 */}
      <Grid columns={3} stackable style={{ marginBottom: '20px' }}>
        <Grid.Column>
          <Statistic>
            <Statistic.Value>{stats.totalGroups}</Statistic.Value>
            <Statistic.Label>总用户组</Statistic.Label>
          </Statistic>
        </Grid.Column>
        <Grid.Column>
          <Statistic>
            <Statistic.Value>{stats.totalUsers}</Statistic.Value>
            <Statistic.Label>总用户数</Statistic.Label>
          </Statistic>
        </Grid.Column>
        <Grid.Column>
          <Statistic>
            <Statistic.Value>{stats.totalQuota.toLocaleString()}</Statistic.Value>
            <Statistic.Label>总配额</Statistic.Label>
          </Statistic>
        </Grid.Column>
      </Grid>

      <Segment>
        <Button
          primary
          icon
          labelPosition="left"
          onClick={openCreateModal}
          style={{ marginBottom: '20px' }}
        >
          <Icon name="plus" />
          {t('group.buttons.add')}
        </Button>

        <Table celled>
          <Table.Header>
            <Table.Row>
              <Table.HeaderCell>{t('group.table.name')}</Table.HeaderCell>
              <Table.HeaderCell>{t('group.table.description')}</Table.HeaderCell>
              <Table.HeaderCell>{t('group.table.user_count')}</Table.HeaderCell>
              <Table.HeaderCell>{t('group.table.quota')}</Table.HeaderCell>
              <Table.HeaderCell>{t('group.table.status')}</Table.HeaderCell>
              <Table.HeaderCell>{t('group.table.actions')}</Table.HeaderCell>
            </Table.Row>
          </Table.Header>

          <Table.Body>
            {groups.map((group, index) => (
              <Table.Row key={group.id || `group-${index}`}>
                <Table.Cell>
                  <Header as="h4">
                    <Header.Content>
                      {group.name}
                      <Header.Subheader>{group.id}</Header.Subheader>
                    </Header.Content>
                  </Header>
                </Table.Cell>
                <Table.Cell>{group.description || '-'}</Table.Cell>
                <Table.Cell>{group.user_count || 0}</Table.Cell>
                <Table.Cell>{group.total_quota ? group.total_quota.toLocaleString() : '-'}</Table.Cell>
                <Table.Cell>{renderGroupStatus(group)}</Table.Cell>
                <Table.Cell>
                  <Button.Group size="mini">
                    <Button
                      icon
                      onClick={() => openEditModal(group)}
                      title={t('group.buttons.edit')}
                    >
                      <Icon name="edit" />
                    </Button>
                    <Button
                      icon
                      negative
                      onClick={() => {
                        setEditingGroup(group);
                        setConfirmOpen(true);
                      }}
                      title={t('group.buttons.delete')}
                      disabled={group.user_count > 0}
                    >
                      <Icon name="trash" />
                    </Button>
                  </Button.Group>
                </Table.Cell>
              </Table.Row>
            ))}
          </Table.Body>
        </Table>

        {groups.length === 0 && !loading && (
          <Message info>
            <Message.Header>暂无用户组</Message.Header>
            <p>点击上方按钮创建第一个用户组</p>
          </Message>
        )}
      </Segment>

      {/* 创建/编辑模态框 */}
      <Modal open={modalOpen} onClose={() => setModalOpen(false)} size="small">
        <Modal.Header>
          {editingGroup ? t('group.edit.title_edit') : t('group.edit.title_create')}
        </Modal.Header>
        <Modal.Content>
          <Form>
            <Form.Field>
              <label>{t('group.edit.name')}</label>
              <Form.Input
                placeholder={t('group.edit.name_placeholder')}
                value={formData.name}
                onChange={(e, { value }) => setFormData({ ...formData, name: value })}
              />
            </Form.Field>
            <Form.Field>
              <label>{t('group.edit.description')}</label>
              <Form.TextArea
                placeholder={t('group.edit.description_placeholder')}
                value={formData.description}
                onChange={(e, { value }) => setFormData({ ...formData, description: value })}
              />
            </Form.Field>
            <Form.Field>
              <label>{t('group.edit.quota')}</label>
              <Form.Input
                type="number"
                placeholder={t('group.edit.quota_placeholder')}
                value={formData.total_quota}
                onChange={(e, { value }) => setFormData({ ...formData, total_quota: parseInt(value) || 0 })}
              />
            </Form.Field>
          </Form>
        </Modal.Content>
        <Modal.Actions>
          <Button onClick={() => setModalOpen(false)}>
            {t('common.cancel')}
          </Button>
          <Button primary onClick={handleSubmit}>
            {editingGroup ? t('common.update') : t('common.create')}
          </Button>
        </Modal.Actions>
      </Modal>

      {/* 删除确认框 */}
      <Confirm
        open={confirmOpen}
        header={t('group.buttons.confirm_delete')}
        content={`确定要删除用户组 "${editingGroup?.name}" 吗？此操作不可撤销。`}
        onCancel={() => setConfirmOpen(false)}
        onConfirm={handleDelete}
        cancelButton={t('common.cancel')}
        confirmButton={t('common.delete')}
      />
    </Container>
  );
};

export default GroupManagement; 