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
  const [showGroupModal, setShowGroupModal] = useState(false);
  const [selectedToken, setSelectedToken] = useState(null);
  const [selectedGroup, setSelectedGroup] = useState('');
  const [updatingGroup, setUpdatingGroup] = useState(false);

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
        // 令牌数据现在已经包含group字段，无需额外请求
        const tokensWithGroups = data.map(token => ({
          ...token,
          group: token.group || 'default'
        }));

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

  // 获取令牌的用户组
  const getTokenGroup = async (tokenId) => {
    try {
      const response = await API.get(`/api/token/${tokenId}`);
      if (response.data.success && response.data.data) {
        // 直接从令牌数据中获取用户组
        return response.data.data.group || 'default';
      }
      return 'default';
    } catch (error) {
      console.error('获取令牌用户组失败:', error);
      return 'default';
    }
  };

  // 更新令牌用户组
  const updateTokenGroup = async () => {
    if (!selectedToken || !selectedGroup) {
      showError('请选择用户组');
      return;
    }

    try {
      setUpdatingGroup(true);
      const response = await API.put(`/api/token/${selectedToken.id}/group`, {
        group: selectedGroup,
      });
      
      if (response.data.success) {
        showSuccess('令牌用户组更新成功');
        setShowGroupModal(false);
        setSelectedToken(null);
        setSelectedGroup('');
        
        // 立即更新当前令牌的用户组显示
        const updatedTokens = tokens.map(token => {
          if (token.id === selectedToken.id) {
            return { ...token, group: selectedGroup };
          }
          return token;
        });
        setTokens(updatedTokens);
        
        // 同时刷新整个列表以确保数据一致性
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
      setUpdatingGroup(false);
    }
  };

  // 打开用户组修改模态框
  const openGroupModal = async (token) => {
    try {
      setSelectedToken(token);
      // 优先使用令牌中已有的用户组信息，避免重复请求
      const currentGroup = token.group || await getTokenGroup(token.id);
      setSelectedGroup(currentGroup);
      setShowGroupModal(true);
    } catch (error) {
      console.error('打开用户组模态框失败:', error);
      showError('打开用户组模态框失败');
    }
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
      setTokens(data);
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

      <Table basic={'very'} compact size='small'>
        <Table.Header>
          <Table.Row>
            <Table.HeaderCell
              style={{ cursor: 'pointer' }}
              onClick={() => {
                sortToken('name');
              }}
            >
              {t('token.table.name')}
            </Table.HeaderCell>
            <Table.HeaderCell
              style={{ cursor: 'pointer' }}
              onClick={() => {
                sortToken('status');
              }}
            >
              {t('token.table.status')}
            </Table.HeaderCell>
            <Table.HeaderCell
              style={{ cursor: 'pointer' }}
              onClick={() => {
                sortToken('used_quota');
              }}
            >
              {t('token.table.used_quota')}
            </Table.HeaderCell>
            <Table.HeaderCell
              style={{ cursor: 'pointer' }}
              onClick={() => {
                sortToken('remain_quota');
              }}
            >
              {t('token.table.remain_quota')}
            </Table.HeaderCell>
            <Table.HeaderCell
              style={{ cursor: 'pointer' }}
              onClick={() => {
                sortToken('created_time');
              }}
            >
              {t('token.table.created_time')}
            </Table.HeaderCell>
            <Table.HeaderCell
              style={{ cursor: 'pointer' }}
              onClick={() => {
                sortToken('expired_time');
              }}
            >
              {t('token.table.expired_time')}
            </Table.HeaderCell>
            <Table.HeaderCell>{t('token.table.user_group')}</Table.HeaderCell>
            <Table.HeaderCell>{t('token.table.actions')}</Table.HeaderCell>
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
                  <Table.Cell>
                    {token.name ? token.name : t('token.table.no_name')}
                  </Table.Cell>
                  <Table.Cell>{renderStatus(token.status, t)}</Table.Cell>
                  <Table.Cell>{renderQuota(token.used_quota, t)}</Table.Cell>
                  <Table.Cell>
                    {token.unlimited_quota
                      ? t('token.table.unlimited')
                      : renderQuota(token.remain_quota, t, 2)}
                  </Table.Cell>
                  <Table.Cell>{renderTimestamp(token.created_time)}</Table.Cell>
                  <Table.Cell>
                    {token.expired_time === -1
                      ? t('token.table.never_expire')
                      : renderTimestamp(token.expired_time)}
                  </Table.Cell>
                  <Table.Cell>
                    <Label basic color='blue'>
                      {token.group || 'default'}
                    </Label>
                    <Button
                      size='mini'
                      icon
                      basic
                      onClick={() => openGroupModal(token)}
                      style={{ marginLeft: '5px' }}
                    >
                      <Icon name='edit' />
                    </Button>
                  </Table.Cell>
                  <Table.Cell>
                    <div>
                      <Button.Group color='green' size={'tiny'}>
                        <Button
                          size={'tiny'}
                          positive
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
                      </Button.Group>{' '}
                      <Button.Group color='olive' size={'tiny'}>
                        <Button
                          size={'tiny'}
                          positive
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
                      </Button.Group>{' '}
                      <Popup
                        trigger={
                          <Button size='mini' negative>
                            {t('token.buttons.delete')}
                          </Button>
                        }
                        on='click'
                        flowing
                        hoverable
                      >
                        <Button
                          size={'tiny'}
                          negative
                          onClick={() => {
                            manageToken(token.id, 'delete', idx);
                          }}
                        >
                          {t('token.buttons.confirm_delete')} {token.name}
                        </Button>
                      </Popup>
                      <Button
                        size={'tiny'}
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
                        size={'tiny'}
                        as={Link}
                        to={'/token/edit/' + token.id}
                      >
                        {t('token.buttons.edit')}
                      </Button>
                    </div>
                  </Table.Cell>
                </Table.Row>
              );
            })}
        </Table.Body>

        <Table.Footer>
          <Table.Row>
            <Table.HeaderCell colSpan='8'>
              <Button size='small' as={Link} to='/token/add' loading={loading}>
                {t('token.buttons.add')}
              </Button>
              <Button size='small' onClick={refresh} loading={loading}>
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
                size='small'
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

      {/* 用户组修改模态框 */}
      <Modal
        open={showGroupModal}
        onClose={() => setShowGroupModal(false)}
        size='small'
      >
        <Header icon='users' content='修改令牌用户组' />
        <Modal.Content>
          <p>
            令牌: <strong>{selectedToken?.name || selectedToken?.key}</strong>
          </p>
          <Form>
            <Form.Field>
              <label>用户组</label>
              <Dropdown
                placeholder='选择用户组'
                fluid
                selection
                options={groups.map(group => ({
                  key: group.name,
                  text: group.name,
                  value: group.name
                }))}
                value={selectedGroup}
                onChange={(e, { value }) => setSelectedGroup(value)}
              />
            </Form.Field>
          </Form>
        </Modal.Content>
        <Modal.Actions>
          <Button onClick={() => setShowGroupModal(false)}>
            取消
          </Button>
          <Button
            color='blue'
            onClick={updateTokenGroup}
            loading={updatingGroup}
            disabled={!selectedGroup}
          >
            确认修改
          </Button>
        </Modal.Actions>
      </Modal>
    </>
  );
};

export default TokensTable;
