import { showError } from './utils';
import { authHeader } from './auth-header';
import axios from 'axios';

export const API = axios.create({
  baseURL: process.env.REACT_APP_SERVER ? process.env.REACT_APP_SERVER : '',
});

// 请求拦截器：自动添加认证头
API.interceptors.request.use(
  (config) => {
    const headers = authHeader();
    if (headers.Authorization) {
      config.headers.Authorization = headers.Authorization;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// 响应拦截器：处理错误
API.interceptors.response.use(
  (response) => response,
  (error) => {
    // 不在拦截器中显示错误，让组件自己处理
    // 只处理网络级别的错误
    if (!error.response) {
      showError('网络错误，请稍后重试');
    }
    return Promise.reject(error);
  }
);
