import React, { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Card, Button, Icon, Message } from 'semantic-ui-react';
import { API, showError, showSuccess, isAdmin } from '../../helpers';
import { marked } from 'marked';

const About = () => {
  const { t } = useTranslation();
  const [about, setAbout] = useState('');
  const [aboutLoaded, setAboutLoaded] = useState(false);
  const [showCoursewareButton, setShowCoursewareButton] = useState(false);
  const [clickCount, setClickCount] = useState(0);
  const [lastClickTime, setLastClickTime] = useState(0);

  const displayAbout = async () => {
    setAbout(localStorage.getItem('about') || '');
    const res = await API.get('/api/about');
    const { success, message, data } = res.data;
    if (success) {
      let aboutContent = data;
      if (!data.startsWith('https://')) {
        aboutContent = marked.parse(data);
      }
      setAbout(aboutContent);
      localStorage.setItem('about', aboutContent);
    } else {
      showError(message);
      setAbout(t('about.loading_failed'));
    }
    setAboutLoaded(true);
  };

  // 检查是否已开启课件平台集成功能
  const checkCoursewareEnabled = () => {
    const enabled = localStorage.getItem('courseware_integration_enabled') === 'true';
    setShowCoursewareButton(enabled);
  };

  // 秘籍操作：连续点击5次开启课件平台集成
  const handleSecretClick = () => {
    const now = Date.now();
    if (now - lastClickTime < 2000) { // 2秒内的点击才有效
      const newCount = clickCount + 1;
      setClickCount(newCount);
      setLastClickTime(now);
      
      if (newCount >= 5) {
        // 开启课件平台集成功能
        localStorage.setItem('courseware_integration_enabled', 'true');
        setShowCoursewareButton(true);
        setClickCount(0);
        showSuccess('秘籍激活成功！课件平台集成功能已开启');
      }
    } else {
      // 重置计数
      setClickCount(1);
      setLastClickTime(now);
    }
  };

  // 跳转到课件平台集成页面
  const goToCourseware = () => {
    window.location.href = '/courseware';
  };

  useEffect(() => {
    displayAbout().then();
    checkCoursewareEnabled();
  }, []);

  return (
    <>
      {aboutLoaded && about === '' ? (
        <div className='dashboard-container'>
          <Card fluid className='chart-card'>
            <Card.Content>
              <Card.Header className='header'>{t('about.title')}</Card.Header>
              <p>{t('about.description')}</p>
              {t('about.repository')}
              <a href='https://github.com/songquanpeng/one-api'>
                https://github.com/songquanpeng/one-api
              </a>
              
              {/* 秘籍操作区域 */}
              {isAdmin() && (
                <div style={{ marginTop: '20px', textAlign: 'center' }}>
                  <div 
                    style={{ 
                      cursor: 'pointer', 
                      padding: '10px', 
                      border: '1px dashed #ccc',
                      borderRadius: '5px',
                      display: 'inline-block',
                      userSelect: 'none'
                    }}
                    onClick={handleSecretClick}
                    title="连续点击5次激活秘籍"
                  >
                    <Icon name="question circle" />
                    <span style={{ marginLeft: '5px', fontSize: '12px', color: '#999' }}>
                      点击次数: {clickCount}/5
                    </span>
                  </div>
                  
                  {showCoursewareButton && (
                    <div style={{ marginTop: '10px' }}>
                      <Button 
                        primary 
                        icon 
                        labelPosition='left'
                        onClick={goToCourseware}
                      >
                        <Icon name='cloud' />
                        课件平台集成
                      </Button>
                      <Message info size='tiny' style={{ marginTop: '10px' }}>
                        <Message.Header>秘籍已激活</Message.Header>
                        <p>课件平台集成功能已开启，点击上方按钮进入管理页面。</p>
                      </Message>
                    </div>
                  )}
                </div>
              )}
            </Card.Content>
          </Card>
        </div>
      ) : (
        <>
          {about.startsWith('https://') ? (
            <iframe
              src={about}
              style={{ width: '100%', height: '100vh', border: 'none' }}
            />
          ) : (
            <div className='dashboard-container'>
              <Card fluid className='chart-card'>
                <Card.Content>
                  <div
                    style={{ fontSize: 'larger' }}
                    dangerouslySetInnerHTML={{ __html: about }}
                  ></div>
                  
                  {/* 秘籍操作区域 - 在自定义关于页面中也显示 */}
                  {isAdmin() && (
                    <div style={{ marginTop: '20px', textAlign: 'center', borderTop: '1px solid #eee', paddingTop: '20px' }}>
                      <div 
                        style={{ 
                          cursor: 'pointer', 
                          padding: '10px', 
                          border: '1px dashed #ccc',
                          borderRadius: '5px',
                          display: 'inline-block',
                          userSelect: 'none'
                        }}
                        onClick={handleSecretClick}
                        title="连续点击5次激活秘籍"
                      >
                        <Icon name="question circle" />
                        <span style={{ marginLeft: '5px', fontSize: '12px', color: '#999' }}>
                          点击次数: {clickCount}/5
                        </span>
                      </div>
                      
                      {showCoursewareButton && (
                        <div style={{ marginTop: '10px' }}>
                          <Button 
                            primary 
                            icon 
                            labelPosition='left'
                            onClick={goToCourseware}
                          >
                            <Icon name='cloud' />
                            课件平台集成
                          </Button>
                          <Message info size='tiny' style={{ marginTop: '10px' }}>
                            <Message.Header>秘籍已激活</Message.Header>
                            <p>课件平台集成功能已开启，点击上方按钮进入管理页面。</p>
                          </Message>
                        </div>
                      )}
                    </div>
                  )}
                </Card.Content>
              </Card>
            </div>
          )}
        </>
      )}
    </>
  );
};

export default About;
