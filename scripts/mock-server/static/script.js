// 全局变量
let currentEditUserId = null;

// 页面加载完成后初始化
document.addEventListener('DOMContentLoaded', function() {
    initializeNavigation();
    loadUsers();
    loadConfig();
});

// 初始化导航
function initializeNavigation() {
    const navLinks = document.querySelectorAll('.nav-link');
    const panels = document.querySelectorAll('.panel');
    
    navLinks.forEach(link => {
        link.addEventListener('click', function(e) {
            e.preventDefault();
            
            // 移除所有活动状态
            navLinks.forEach(l => l.classList.remove('active'));
            panels.forEach(p => p.classList.remove('active'));
            
            // 添加活动状态
            this.classList.add('active');
            const targetId = this.getAttribute('href').substring(1);
            document.getElementById(targetId).classList.add('active');
        });
    });
}

// 加载用户列表
async function loadUsers() {
    try {
        const response = await fetch('/web/users');
        const users = await response.json();
        displayUsers(users);
    } catch (error) {
        showMessage('加载用户列表失败: ' + error.message, 'error');
    }
}

// 显示用户列表
function displayUsers(users) {
    const tbody = document.getElementById('userTableBody');
    tbody.innerHTML = '';
    
    users.forEach(user => {
        const row = document.createElement('tr');
        row.innerHTML = `
            <td>${user.teacher_id}</td>
            <td>${user.teacher_name}</td>
            <td>${user.school_name}</td>
            <td>${user.subject_name}</td>
            <td>${user.oneapi_group}</td>
            <td>${user.preferred_model}</td>
            <td>
                <button class="btn btn-primary" onclick="editUser('${user.teacher_id}')">编辑</button>
                <button class="btn btn-danger" onclick="deleteUser('${user.teacher_id}')">删除</button>
            </td>
        `;
        tbody.appendChild(row);
    });
}

// 编辑用户
async function editUser(teacherId) {
    try {
        const response = await fetch(`/web/users/${teacherId}`);
        const user = await response.json();
        
        // 填充表单
        document.getElementById('teacherId').value = user.teacher_id;
        document.getElementById('teacherName').value = user.teacher_name;
        document.getElementById('schoolId').value = user.school_id;
        document.getElementById('schoolName').value = user.school_name;
        document.getElementById('subjectId').value = user.subject_id;
        document.getElementById('subjectName').value = user.subject_name;
        document.getElementById('oneapiGroup').value = user.oneapi_group;
        document.getElementById('preferredModel').value = user.preferred_model;
        
        currentEditUserId = teacherId;
        
        // 切换到用户管理面板
        document.querySelector('a[href="#users"]').click();
        
        showMessage('用户信息已加载到表单中', 'info');
    } catch (error) {
        showMessage('加载用户信息失败: ' + error.message, 'error');
    }
}

// 删除用户
async function deleteUser(teacherId) {
    if (!confirm('确定要删除这个用户吗？')) {
        return;
    }
    
    try {
        const response = await fetch(`/web/users/${teacherId}`, {
            method: 'DELETE'
        });
        
        if (response.ok) {
            showMessage('用户删除成功', 'success');
            loadUsers();
        } else {
            const error = await response.json();
            showMessage('删除用户失败: ' + error.error, 'error');
        }
    } catch (error) {
        showMessage('删除用户失败: ' + error.message, 'error');
    }
}

// 清空表单
function clearForm() {
    document.getElementById('userForm').reset();
    currentEditUserId = null;
}

// 用户表单提交
document.getElementById('userForm').addEventListener('submit', async function(e) {
    e.preventDefault();
    
    const formData = new FormData(this);
    const userData = {
        teacher_id: formData.get('teacherId'),
        teacher_name: formData.get('teacherName'),
        school_id: parseInt(formData.get('schoolId')),
        school_name: formData.get('schoolName'),
        subject_id: parseInt(formData.get('subjectId')),
        subject_name: formData.get('subjectName'),
        oneapi_group: formData.get('oneapiGroup'),
        preferred_model: formData.get('preferredModel')
    };
    
    try {
        const url = currentEditUserId ? '/web/users' : '/web/users';
        const method = currentEditUserId ? 'PUT' : 'POST';
        
        const response = await fetch(url, {
            method: method,
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(userData)
        });
        
        if (response.ok) {
            const result = await response.json();
            showMessage(result.message, 'success');
            clearForm();
            loadUsers();
        } else {
            const error = await response.json();
            showMessage('保存用户失败: ' + error.error, 'error');
        }
    } catch (error) {
        showMessage('保存用户失败: ' + error.message, 'error');
    }
});

// 加载配置
async function loadConfig() {
    try {
        const response = await fetch('/web/config');
        const config = await response.json();
        
        document.getElementById('serverPort').value = config.server.port;
        document.getElementById('serverHost').value = config.server.host;
        document.getElementById('apiKey').value = config.api.api_key;
        document.getElementById('responseDelay').value = config.api.response_delay;
        document.getElementById('errorRate').value = config.api.error_rate;
    } catch (error) {
        showMessage('加载配置失败: ' + error.message, 'error');
    }
}

// 配置表单提交
document.getElementById('configForm').addEventListener('submit', async function(e) {
    e.preventDefault();
    
    const formData = new FormData(this);
    const configData = {
        port: parseInt(formData.get('port')),
        host: formData.get('host'),
        api_key: formData.get('apiKey'),
        response_delay: formData.get('responseDelay'),
        error_rate: parseFloat(formData.get('errorRate'))
    };
    
    try {
        const response = await fetch('/web/config', {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(configData)
        });
        
        if (response.ok) {
            const result = await response.json();
            showMessage(result.message, 'success');
        } else {
            const error = await response.json();
            showMessage('保存配置失败: ' + error.error, 'error');
        }
    } catch (error) {
        showMessage('保存配置失败: ' + error.message, 'error');
    }
});

// 重置为默认
async function resetToDefault() {
    if (!confirm('确定要重置为默认数据吗？这将清空所有用户数据。')) {
        return;
    }
    
    try {
        const response = await fetch('/web/reset', {
            method: 'POST'
        });
        
        if (response.ok) {
            const result = await response.json();
            showMessage(result.message, 'success');
            loadUsers();
            loadConfig();
        } else {
            const error = await response.json();
            showMessage('重置失败: ' + error.error, 'error');
        }
    } catch (error) {
        showMessage('重置失败: ' + error.message, 'error');
    }
}

// 测试API
async function testAPI(type) {
    const teacherId = document.getElementById('testTeacherId').value;
    const resultElement = document.getElementById(`testResult${type === 'teacher_info' ? '1' : type === 'teacher_ids' ? '2' : '3'}`);
    
    try {
        const url = `/web/test/${type}${teacherId ? `?teacher_id=${teacherId}` : ''}`;
        const response = await fetch(url);
        const result = await response.json();
        
        resultElement.textContent = JSON.stringify(result, null, 2);
        resultElement.style.color = response.ok ? '#27ae60' : '#e74c3c';
    } catch (error) {
        resultElement.textContent = '测试失败: ' + error.message;
        resultElement.style.color = '#e74c3c';
    }
}

// 测试批量API
async function testBatchAPI() {
    const teacherIds = document.getElementById('testBatchIds').value;
    const resultElement = document.getElementById('testResult3');
    
    if (!teacherIds) {
        resultElement.textContent = '请输入教师ID列表';
        resultElement.style.color = '#e74c3c';
        return;
    }
    
    try {
        const response = await fetch('/web/test/batch?teacher_id=' + teacherIds);
        const result = await response.json();
        
        resultElement.textContent = JSON.stringify(result, null, 2);
        resultElement.style.color = response.ok ? '#27ae60' : '#e74c3c';
    } catch (error) {
        resultElement.textContent = '测试失败: ' + error.message;
        resultElement.style.color = '#e74c3c';
    }
}

// 显示消息
function showMessage(message, type) {
    // 创建消息元素
    const messageDiv = document.createElement('div');
    messageDiv.className = type;
    messageDiv.textContent = message;
    
    // 插入到页面顶部
    const container = document.querySelector('.container');
    container.insertBefore(messageDiv, container.firstChild);
    
    // 3秒后自动移除
    setTimeout(() => {
        messageDiv.remove();
    }, 3000);
} 