import React, { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import {
  Button,
  Dropdown,
  Form,
  Label,
  Pagination,
  Popup,
  Table,
  Modal,
  Header,
  Icon,
} from 'semantic-ui-react';
import { Link } from 'react-router-dom';
import {
  API,
  copy,
  showError,
  showSuccess,
  showWarning,
  timestamp2string,
} from '../helpers';

import { ITEMS_PER_PAGE } from '../constants';
import { renderQuota } from '../helpers/render';

function renderTimestamp(timestamp) {
  return <>{timestamp2string(timestamp)}</>;
}

function renderStatus(status, t) {
  switch (status) {
    case 1:
      return (
        <Label basic color='green'>
          {t('token.table.status_enabled')}
        </Label>
      );
    case 2:
      return (
        <Label basic color='red'>
          {t('token.table.status_disabled')}
        </Label>
      );
    case 3:
      return (
        <Label basic color='yellow'>
          {t('token.table.status_expired')}
        </Label>
      );
    case 4:
      return (
        <Label basic color='grey'>
          {t('token.table.status_depleted')}
        </Label>
      );
    default:
      return (
        <Label basic color='black'>
          {t('token.table.status_unknown')}
        </Label>
      );
  }
}

const TokensTable = () => {
  const { t } = useTranslation();

  const COPY_OPTIONS = [
    { key: 'raw', text: t('token.copy_options.raw'), value: '' },
    { key: 'next', text: t('token.copy_options.next'), value: 'next' },
    { key: 'ama', text: t('token.copy_options.ama'), value: 'ama' },
    { key: 'opencat', text: t('token.copy_options.opencat'), value: 'opencat' },
    { key: 'lobe', text: t('token.copy_options.lobe'), value: 'lobechat' },
  ];

  const OPEN_LINK_OPTIONS = [
    { key: 'next', text: t('token.copy_options.next'), value: 'next' },
    { key: 'ama', text: t('token.copy_options.ama'), value: 'ama' },
    { key: 'opencat', text: t('token.copy_options.opencat'), value: 'opencat' },
    { key: 'lobe', text: t('token.copy_options.lobe'), value: 'lobechat' },
  ];

  const [tokens, setTokens] = useState([]);
  const [loading, setLoading] = useState(true);
  const [activePage, setActivePage] = useState(1);
  const [searchKeyword, setSearchKeyword] = useState('');
  const [searching, setSearching] = useState(false);
  const [showTopUpModal, setShowTopUpModal] = useState(false);
  const [targetTokenIdx, setTargetTokenIdx] = useState(0);
  const [orderBy, setOrderBy] = useState('');
  
  // 用户组相关状态
  const [groups, setGroups] = useState([]);
  const [selectedToken, setSelectedToken] = useState(null);
  
  // 多用户组相关状态
  const [showMultiGroupModal, setShowMultiGroupModal] = useState(false);
  const [selectedTokenGroups, setSelectedTokenGroups] = useState([]);
  const [updatingMultiGroups, setUpdatingMultiGroups] = useState(false);

  // 加载用户组列表
  useEffect(() => {
    loadGroups();
  }, []);

  // 加载令牌列表
  useEffect(() => {
    loadTokens(0);
  }, [orderBy]);

  const loadTokens = async (startIdx) => {
    try {
      const res = await API.get(`/api/token/?p=${startIdx}&order=${orderBy}`);
      const { success, message, data } = res.data;
      if (success) {
        // 令牌数据现在已经包含 user_groups 字段，需兼容字符串或数组两种形态
        const tokensWithGroups = data.map((token) => {
          let groupsArr = [];
          if (Array.isArray(token.user_groups)) {
            groupsArr = token.user_groups;
          } else if (typeof token.user_groups === 'string') {
            try {
              const parsed = JSON.parse(token.user_groups);
              if (Array.isArray(parsed)) groupsArr = parsed;
            } catch (e) {
              // ignore parse error, fallback to empty
            }
          }
          const group = groupsArr.length > 0 ? groupsArr[0] : 'default';
          return { ...token, group };
        });

        if (startIdx === 0) {
          setTokens(tokensWithGroups);
        } else {
          let newTokens = [...tokens];
          const startIndex = startIdx * ITEMS_PER_PAGE;
          newTokens.splice(startIndex, ITEMS_PER_PAGE, ...tokensWithGroups);
          setTokens(newTokens);
        }
      } else {
        showError(message);
      }
    } catch (error) {
      console.error('加载令牌失败:', error);
      showError('加载令牌失败');
    } finally {
      setLoading(false);
    }
  };

  // 加载用户组列表
  const loadGroups = async () => {
    try {
      const response = await API.get('/api/group/detail');
      if (response.data.success) {
        const groupData = response.data.data || [];
        setGroups(groupData);
      } else {
        console.error('加载用户组失败:', response.data.message);
        setGroups([]);
      }
    } catch (error) {
      console.error('加载用户组失败:', error.message);
      setGroups([]);
    }
  };







  // 获取令牌的多用户组配置
  const getTokenGroups = async (tokenId) => {
    try {
      const response = await API.get(`/api/token/${tokenId}/groups`);
      if (response.data.success) {
        let groups = response.data.data.user_groups;
        if (typeof groups === 'string') {
          try {
            groups = JSON.parse(groups);
          } catch (e) {
            groups = [];
          }
        }
        if (!Array.isArray(groups) || groups.length === 0) {
          groups = ['default'];
        }
        return groups;
      } else {
        console.error('获取令牌用户组失败:', response.data.message);
        return ['default'];
      }
    } catch (error) {
      console.error('获取令牌用户组失败:', error);
      return ['default'];
    }
  };

  // 打开多用户组配置模态框
  const openMultiGroupModal = async (token) => {
    try {
      setSelectedToken(token);
      // 获取当前令牌的用户组列表
      const tokenGroups = await getTokenGroups(token.id);
      setSelectedTokenGroups(tokenGroups);
      setShowMultiGroupModal(true);
    } catch (error) {
      console.error('打开多用户组模态框失败:', error);
      showError('打开多用户组模态框失败');
    }
  };

  // 更新令牌的多用户组配置
  const updateTokenGroups = async () => {
    if (!selectedToken || selectedTokenGroups.length === 0) {
      showError('请至少选择一个用户组');
      return;
    }

    try {
      setUpdatingMultiGroups(true);
      const response = await API.put(`/api/token/${selectedToken.id}/groups`, {
        user_groups: selectedTokenGroups,
      });

      if (response.data.success) {
        showSuccess('令牌用户组配置更新成功');
        setShowMultiGroupModal(false);
        setSelectedToken(null);
        setSelectedTokenGroups([]);
        
        // 刷新令牌列表
        await loadTokens(0);
      } else {
        showError(response.data.message || '更新失败');
      }
    } catch (error) {
      if (error.response && error.response.data && error.response.data.message) {
        showError(error.response.data.message);
      } else {
        showError('更新失败: ' + error.message);
      }
    } finally {
      setUpdatingMultiGroups(false);
    }
  };

  // 添加用户组到列表
  const addGroupToToken = (groupName) => {
    if (groupName && !selectedTokenGroups.includes(groupName)) {
      setSelectedTokenGroups([...selectedTokenGroups, groupName]);
    }
  };

  // 从列表中移除用户组
  const removeGroupFromToken = (groupName) => {
    setSelectedTokenGroups(selectedTokenGroups.filter(g => g !== groupName));
  };

  // 设为默认用户组（放到列表首位）
  const setDefaultGroup = (groupName) => {
    const filtered = selectedTokenGroups.filter((g) => g !== groupName);
    setSelectedTokenGroups([groupName, ...filtered]);
  };

  const onPaginationChange = (e, { activePage }) => {
    (async () => {
      if (activePage === Math.ceil(tokens.length / ITEMS_PER_PAGE) + 1) {
        // In this case we have to load more data and then append them.
        await loadTokens(activePage - 1);
      }
      setActivePage(activePage);
    })();
  };

  const refresh = async () => {
    setLoading(true);
    await loadTokens(activePage - 1);
  };

  const onCopy = async (type, key) => {
    let status = localStorage.getItem('status');
    let serverAddress = '';
    if (status) {
      status = JSON.parse(status);
      serverAddress = status.server_address;
    }
    if (serverAddress === '') {
      serverAddress = window.location.origin;
    }
    let encodedServerAddress = encodeURIComponent(serverAddress);
    const nextLink = localStorage.getItem('chat_link');
    let nextUrl;

    if (nextLink) {
      nextUrl =
        nextLink + `/#/?settings={"key":"sk-${key}","url":"${serverAddress}"}`;
    } else {
      nextUrl = `https://app.nextchat.dev/#/?settings={"key":"sk-${key}","url":"${serverAddress}"}`;
    }

    let url;
    switch (type) {
      case 'ama':
        url = `ama://set-api-key?server=${encodedServerAddress}&key=sk-${key}`;
        break;
      case 'opencat':
        url = `opencat://team/join?domain=${encodedServerAddress}&token=sk-${key}`;
        break;
      case 'next':
        url = nextUrl;
        break;
      case 'lobechat':
        url =
          nextLink +
          `/?settings={"keyVaults":{"openai":{"apiKey":"sk-${key}","baseURL":"${serverAddress}/v1"}}}`;
        break;
      default:
        url = `sk-${key}`;
    }
    if (await copy(url)) {
      showSuccess(t('token.messages.copy_success'));
    } else {
      showWarning(t('token.messages.copy_failed'));
      setSearchKeyword(url);
    }
  };

  const onOpenLink = async (type, key) => {
    let status = localStorage.getItem('status');
    let serverAddress = '';
    if (status) {
      status = JSON.parse(status);
      serverAddress = status.server_address;
    }
    if (serverAddress === '') {
      serverAddress = window.location.origin;
    }
    let encodedServerAddress = encodeURIComponent(serverAddress);
    const chatLink = localStorage.getItem('chat_link');
    let defaultUrl;

    if (chatLink) {
      defaultUrl =
        chatLink + `/#/?settings={"key":"sk-${key}","url":"${serverAddress}"}`;
    } else {
      defaultUrl = `https://app.nextchat.dev/#/?settings={"key":"sk-${key}","url":"${serverAddress}"}`;
    }
    let url;
    switch (type) {
      case 'ama':
        url = `ama://set-api-key?server=${encodedServerAddress}&key=sk-${key}`;
        break;

      case 'opencat':
        url = `opencat://team/join?domain=${encodedServerAddress}&token=sk-${key}`;
        break;

      case 'lobechat':
        url =
          chatLink +
          `/?settings={"keyVaults":{"openai":{"apiKey":"sk-${key}","baseURL":"${serverAddress}/v1"}}}`;
        break;

      default:
        url = defaultUrl;
    }

    window.open(url, '_blank');
  };



  const manageToken = async (id, action, idx) => {
    let data = { id };
    let res;
    switch (action) {
      case 'delete':
        res = await API.delete(`/api/token/${id}/`);
        break;
      case 'enable':
        data.status = 1;
        res = await API.put('/api/token/?status_only=true', data);
        break;
      case 'disable':
        data.status = 2;
        res = await API.put('/api/token/?status_only=true', data);
        break;
    }
    const { success, message } = res.data;
    if (success) {
      showSuccess(t('token.messages.operation_success'));
      let token = res.data.data;
      let newTokens = [...tokens];
      let realIdx = (activePage - 1) * ITEMS_PER_PAGE + idx;
      if (action === 'delete') {
        newTokens[realIdx].deleted = true;
      } else {
        newTokens[realIdx].status = token.status;
      }
      setTokens(newTokens);
    } else {
      showError(message);
    }
  };

  const searchTokens = async () => {
    if (searchKeyword === '') {
      // if keyword is blank, load files instead.
      await loadTokens(0);
      setActivePage(1);
      setOrderBy('');
      return;
    }
    setSearching(true);
    const res = await API.get(`/api/token/search?keyword=${searchKeyword}`);
    const { success, message, data } = res.data;
    if (success) {
      // 与列表加载一致，规范 user_groups 并提取默认组用于显示
      const tokensWithGroups = (data || []).map((token) => {
        let groupsArr = [];
        if (Array.isArray(token.user_groups)) {
          groupsArr = token.user_groups;
        } else if (typeof token.user_groups === 'string') {
          try {
            const parsed = JSON.parse(token.user_groups);
            if (Array.isArray(parsed)) groupsArr = parsed;
          } catch (e) {}
        }
        const group = groupsArr.length > 0 ? groupsArr[0] : 'default';
        return { ...token, group };
      });
      setTokens(tokensWithGroups);
      setActivePage(1);
    } else {
      showError(message);
    }
    setSearching(false);
  };

  const handleKeywordChange = async (e, { value }) => {
    setSearchKeyword(value.trim());
  };

  const sortToken = (key) => {
    if (tokens.length === 0) return;
    setLoading(true);
    let sortedTokens = [...tokens];
    sortedTokens.sort((a, b) => {
      if (!isNaN(a[key])) {
        // If the value is numeric, subtract to sort
        return a[key] - b[key];
      } else {
        // If the value is not numeric, sort as strings
        return ('' + a[key]).localeCompare(b[key]);
      }
    });
    if (sortedTokens[0].id === tokens[0].id) {
      sortedTokens.reverse();
    }
    setTokens(sortedTokens);
    setLoading(false);
  };

  const handleOrderByChange = (e, { value }) => {
    setOrderBy(value);
    setActivePage(1);
  };

  return (
    <>
      <Form onSubmit={searchTokens}>
        <Form.Input
          icon='search'
          fluid
          iconPosition='left'
          placeholder={t('token.search')}
          value={searchKeyword}
          loading={searching}
          onChange={handleKeywordChange}
        />
      </Form>

      <div style={{ overflowX: 'auto', marginTop: '10px' }}>
        <Table basic={'very'} compact size='small' style={{ tableLayout: 'fixed', minWidth: '800px' }}>
        <Table.Header>
          <Table.Row>
            <Table.HeaderCell
              style={{ cursor: 'pointer', width: '15%' }}
              onClick={() => {
                sortToken('name');
              }}
            >
              {t('token.table.name')}
            </Table.HeaderCell>
            <Table.HeaderCell
              style={{ cursor: 'pointer', width: '8%' }}
              onClick={() => {
                sortToken('status');
              }}
            >
              {t('token.table.status')}
            </Table.HeaderCell>
            <Table.HeaderCell
              style={{ cursor: 'pointer', width: '10%' }}
              onClick={() => {
                sortToken('used_quota');
              }}
            >
              {t('token.table.used_quota')}
            </Table.HeaderCell>
            <Table.HeaderCell
              style={{ cursor: 'pointer', width: '10%' }}
              onClick={() => {
                sortToken('remain_quota');
              }}
            >
              {t('token.table.remain_quota')}
            </Table.HeaderCell>
            <Table.HeaderCell
              style={{ cursor: 'pointer', width: '12%' }}
              onClick={() => {
                sortToken('created_time');
              }}
            >
              {t('token.table.created_time')}
            </Table.HeaderCell>
            <Table.HeaderCell
              style={{ cursor: 'pointer', width: '12%' }}
              onClick={() => {
                sortToken('expired_time');
              }}
            >
              {t('token.table.expired_time')}
            </Table.HeaderCell>
            <Table.HeaderCell style={{ width: '13%' }}>{t('token.table.user_group')}</Table.HeaderCell>
            <Table.HeaderCell style={{ width: '20%' }}>{t('token.table.actions')}</Table.HeaderCell>
          </Table.Row>
        </Table.Header>

        <Table.Body>
          {tokens
            .slice(
              (activePage - 1) * ITEMS_PER_PAGE,
              activePage * ITEMS_PER_PAGE
            )
            .map((token, idx) => {
              if (token.deleted) return <></>;

              const copyOptionsWithHandlers = COPY_OPTIONS.map((option) => ({
                ...option,
                onClick: async () => {
                  await onCopy(option.value, token.key);
                },
              }));

              const openLinkOptionsWithHandlers = OPEN_LINK_OPTIONS.map(
                (option) => ({
                  ...option,
                  onClick: async () => {
                    await onOpenLink(option.value, token.key);
                  },
                })
              );

              return (
                <Table.Row key={token.id}>
                  <Table.Cell style={{ overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                    {token.name ? token.name : t('token.table.no_name')}
                  </Table.Cell>
                  <Table.Cell style={{ overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                    {renderStatus(token.status, t)}
                  </Table.Cell>
                  <Table.Cell style={{ overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                    {renderQuota(token.used_quota, t)}
                  </Table.Cell>
                  <Table.Cell style={{ overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                    {token.unlimited_quota
                      ? t('token.table.unlimited')
                      : renderQuota(token.remain_quota, t, 2)}
                  </Table.Cell>
                  <Table.Cell style={{ overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                    {renderTimestamp(token.created_time)}
                  </Table.Cell>
                  <Table.Cell style={{ overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                    {token.expired_time === -1
                      ? t('token.table.never_expire')
                      : renderTimestamp(token.expired_time)}
                  </Table.Cell>
                  <Table.Cell>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '2px', flexWrap: 'nowrap', minWidth: 0 }}>
                      <Label basic color='blue' style={{ flexShrink: 1, minWidth: 0, maxWidth: '100%' }}>
                        <span style={{ 
                          overflow: 'hidden', 
                          textOverflow: 'ellipsis', 
                          whiteSpace: 'nowrap',
                          display: 'block',
                          maxWidth: '100%'
                        }}>
                          {token.group || 'default'}
                        </span>
                      </Label>

                      <Popup
                        trigger={
                          <Button
                            size='mini'
                            icon
                            basic
                            compact
                            color='green'
                            onClick={() => openMultiGroupModal(token)}
                            style={{ flexShrink: 0, padding: '4px' }}
                          >
                            <Icon name='list' />
                          </Button>
                        }
                        content='配置用户组'
                        position='top center'
                        inverted
                      />
                    </div>
                  </Table.Cell>
                  <Table.Cell>
                    <div
                      style={{
                        display: 'flex',
                        flexDirection: 'column',
                        gap: '2px',
                        alignItems: 'flex-start',
                      }}
                    >
                      <div style={{ display: 'flex', gap: '2px', flexWrap: 'wrap' }}>
                        <Button.Group color='green' size={'mini'}>
                          <Button
                            size={'mini'}
                            positive
                            compact
                            onClick={async () => await onCopy('', token.key)}
                          >
                            {t('token.buttons.copy')}
                          </Button>
                          <Dropdown
                            className='button icon'
                            floating
                            options={copyOptionsWithHandlers}
                            trigger={<></>}
                          />
                        </Button.Group>
                        <Button.Group color='olive' size={'mini'}>
                          <Button
                            size={'mini'}
                            positive
                            compact
                            onClick={() => onOpenLink('', token.key)}
                          >
                            {t('token.buttons.chat')}
                          </Button>
                          <Dropdown
                            className='button icon'
                            floating
                            options={openLinkOptionsWithHandlers}
                            trigger={<></>}
                          />
                        </Button.Group>
                      </div>
                      <div style={{ display: 'flex', gap: '2px', flexWrap: 'wrap' }}>
                        <Popup
                          trigger={
                            <Button size='mini' negative compact>
                              {t('token.buttons.delete')}
                            </Button>
                          }
                          on='click'
                          flowing
                          hoverable
                        >
                          <Button
                            size='mini'
                            negative
                            onClick={() => {
                              manageToken(token.id, 'delete', idx);
                            }}
                          >
                            {t('token.buttons.confirm_delete')} {token.name}
                          </Button>
                        </Popup>
                        <Button
                          size='mini'
                          compact
                          onClick={() => {
                            manageToken(
                              token.id,
                              token.status === 1 ? 'disable' : 'enable',
                              idx
                            );
                          }}
                        >
                          {token.status === 1
                            ? t('token.buttons.disable')
                            : t('token.buttons.enable')}
                        </Button>
                        <Button
                          size='mini'
                          compact
                          as={Link}
                          to={'/token/edit/' + token.id}
                        >
                          {t('token.buttons.edit')}
                        </Button>
                      </div>
                    </div>
                  </Table.Cell>
                </Table.Row>
              );
            })}
        </Table.Body>

        <Table.Footer>
          <Table.Row>
            <Table.HeaderCell colSpan='8'>
              <Button size='tiny' as={Link} to='/token/add' loading={loading}>
                {t('token.buttons.add')}
              </Button>
              <Button size='tiny' onClick={refresh} loading={loading}>
                {t('token.buttons.refresh')}
              </Button>
              <Dropdown
                placeholder={t('token.sort.placeholder')}
                selection
                options={[
                  { key: '', text: t('token.sort.default'), value: '' },
                  {
                    key: 'remain_quota',
                    text: t('token.sort.by_remain'),
                    value: 'remain_quota',
                  },
                  {
                    key: 'used_quota',
                    text: t('token.sort.by_used'),
                    value: 'used_quota',
                  },
                ]}
                value={orderBy}
                onChange={handleOrderByChange}
                style={{ marginLeft: '10px' }}
              />
              <Pagination
                floated='right'
                activePage={activePage}
                onPageChange={onPaginationChange}
                size='tiny'
                siblingRange={1}
                totalPages={
                  Math.ceil(tokens.length / ITEMS_PER_PAGE) +
                  (tokens.length % ITEMS_PER_PAGE === 0 ? 1 : 0)
                }
              />
            </Table.HeaderCell>
          </Table.Row>
        </Table.Footer>
      </Table>
      </div>



      {/* 多用户组配置模态框 */}
      <Modal
        open={showMultiGroupModal}
        onClose={() => setShowMultiGroupModal(false)}
        size='small'
      >
        <Header icon='list' content='配置令牌用户组' />
        <Modal.Content>
          <p>
            令牌: <strong>{selectedToken?.name || selectedToken?.key}</strong>
          </p>
          <p style={{ color: '#666', fontSize: '0.9em', marginBottom: '1em' }}>
            请选择一个默认用户组，其他为非默认用户组。当无法匹配请求中的用户组时，将使用默认用户组。
          </p>
          
          <Form>
            <Form.Field>
              <label>添加用户组</label>
              <Dropdown
                placeholder='选择要添加的用户组'
                fluid
                selection
                clearable
                options={groups
                  .filter(group => !selectedTokenGroups.includes(group.name))
                  .map(group => ({
                    key: group.name,
                    text: group.name,
                    value: group.name
                  }))
                }
                onChange={(e, { value }) => {
                  if (value) {
                    addGroupToToken(value);
                  }
                }}
                value={null}
              />
            </Form.Field>
            
            <Form.Field>
              <label>默认与其他用户组</label>
              <div style={{ border: '1px solid #ddd', borderRadius: '4px', padding: '8px', minHeight: '100px' }}>
                {selectedTokenGroups.length === 0 ? (
                  <div style={{ color: '#999', textAlign: 'center', padding: '20px' }}>
                    请至少添加一个用户组
                  </div>
                ) : (
                  <>
                    {/* 默认用户组 */}
                    <div
                      style={{
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'space-between',
                        padding: '8px',
                        margin: '4px 0',
                        backgroundColor: '#f0fff4',
                        borderRadius: '4px',
                        border: '1px solid #c6f6d5'
                      }}
                    >
                      <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                        <Label basic color='green'>
                          {selectedTokenGroups[0]}
                        </Label>
                        <Label basic color='green' size='mini'>默认</Label>
                      </div>
                      <div style={{ display: 'flex', gap: '4px' }}>
                        <Button
                          size='mini'
                          icon
                          basic
                          compact
                          color='red'
                          onClick={() => removeGroupFromToken(selectedTokenGroups[0])}
                        >
                          <Icon name='trash' />
                        </Button>
                      </div>
                    </div>

                    {/* 其他用户组 */}
                    {selectedTokenGroups.slice(1).map((group, index) => (
                      <div
                        key={`${group}-${index}`}
                        style={{
                          display: 'flex',
                          alignItems: 'center',
                          justifyContent: 'space-between',
                          padding: '8px',
                          margin: '4px 0',
                          backgroundColor: '#f8f9fa',
                          borderRadius: '4px',
                          border: '1px solid #e9ecef'
                        }}
                      >
                        <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                          <Label basic color='blue'>
                            {group}
                          </Label>
                        </div>
                        <div style={{ display: 'flex', gap: '4px' }}>
                          <Button
                            size='mini'
                            icon
                            basic
                            compact
                            color='yellow'
                            onClick={() => setDefaultGroup(group)}
                          >
                            <Icon name='star' /> 设为默认
                          </Button>
                          <Button
                            size='mini'
                            icon
                            basic
                            compact
                            color='red'
                            onClick={() => removeGroupFromToken(group)}
                          >
                            <Icon name='trash' />
                          </Button>
                        </div>
                      </div>
                    ))}
                  </>
                )}
              </div>
            </Form.Field>
            
            {selectedTokenGroups.length > 0 && (
              <div style={{ marginTop: '16px', padding: '12px', backgroundColor: '#e8f5e8', borderRadius: '4px' }}>
                <h4 style={{ margin: '0 0 8px 0', color: '#2d5a2d' }}>工作逻辑说明：</h4>
                <ul style={{ margin: 0, paddingLeft: '20px', color: '#2d5a2d' }}>
                  <li>当请求包含用户ID时，系统会解析用户所属的用户组</li>
                  <li>如果解析的用户组在令牌配置的用户组列表中，则使用该用户组</li>
                  <li>否则（未包含或不在列表中），使用默认用户组</li>
                </ul>
              </div>
            )}
          </Form>
        </Modal.Content>
        <Modal.Actions>
          <Button onClick={() => setShowMultiGroupModal(false)}>
            取消
          </Button>
          <Button
            color='blue'
            onClick={updateTokenGroups}
            loading={updatingMultiGroups}
            disabled={selectedTokenGroups.length === 0}
          >
            确认配置
          </Button>
        </Modal.Actions>
      </Modal>
    </>
  );
};

export default TokensTable;
