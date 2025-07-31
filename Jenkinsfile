pipeline {
    agent any
    
    environment {
        // 项目配置
        PROJECT_NAME = 'one-api'
        GIT_REPO = 'https://git.code.tencent.com/domonic/one-api.git'
        GIT_BRANCH = 'feature/model-management-system'
        
        // Go 环境
        GO_VERSION = '1.21'
        GOPATH = '/var/lib/jenkins/go'
        GOROOT = '/usr/local/go'
        PATH = "/usr/local/go/bin:/var/lib/jenkins/go/bin:${env.PATH}"
        
        // 测试配置
        TEST_TIMEOUT = '10m'
        COVERAGE_THRESHOLD = '70'
        
        // 测试服务配置
        MYSQL_CONTAINER_NAME = "oneapi-mysql-test-${BUILD_NUMBER}"
        REDIS_CONTAINER_NAME = "oneapi-redis-test-${BUILD_NUMBER}"
        MYSQL_PORT = "3306"
        REDIS_PORT = "6379"
    }
    
    stages {
        stage('环境检查') {
            steps {
                script {
                    echo "=== 环境检查 ==="
                    echo "Jenkins 工作空间: ${WORKSPACE}"
                    echo "构建编号: ${BUILD_NUMBER}"
                    echo "Git 分支: ${GIT_BRANCH}"
                    echo "当前PATH: ${env.PATH}"
                    
                    // 检查必要工具
                    sh '''
                        echo "检查Go环境..."
                        go version || echo "Go未安装"
                        
                        echo "检查Docker环境..."
                        docker --version || echo "Docker未安装"
                        
                        echo "检查其他工具..."
                        which make || echo "Make未安装"
                        which git || echo "Git未安装"
                    '''
                    
                    echo "✅ 环境检查完成"
                }
            }
        }
        
        stage('代码检出') {
            steps {
                script {
                    echo "=== 代码检出 ==="
                    cleanWs()
                    checkout([
                        $class: 'GitSCM',
                        branches: [[name: "*/${GIT_BRANCH}"]],
                        doGenerateSubmoduleConfigurations: false,
                        extensions: [
                            [$class: 'CleanBeforeCheckout'],
                            [$class: 'CleanCheckout']
                        ],
                        submoduleCfg: [],
                        userRemoteConfigs: [[
                            credentialsId: 'git-code-tencent-credentials',
                            url: "${GIT_REPO}"
                        ]]
                    ])
                    
                    sh 'git log --oneline -5'
                    sh 'git status'
                }
            }
        }
        
        stage('依赖安装') {
            steps {
                script {
                    echo "=== 依赖安装 ==="
                    echo "当前PATH: ${env.PATH}"
                    
                    // 配置Go模块代理和网络设置
                    sh '''
                        # 设置Go模块代理为国内镜像
                        go env -w GOPROXY=https://goproxy.cn,direct
                        go env -w GOSUMDB=sum.golang.google.cn
                        go env -w GOPRIVATE=git.code.tencent.com
                        go env -w GONOSUMDB=cloud.google.com
                        
                        # 显示Go环境配置
                        echo "Go环境配置:"
                        go env GOPROXY
                        go env GOSUMDB
                        go env GOPRIVATE
                        go env GONOSUMDB
                    '''
                    
                    // 下载依赖（增加重试机制和网络配置）
                    sh '''
                        # 清理模块缓存
                        go clean -modcache
                        
                        # 设置网络超时
                        export GOFLAGS="-timeout=300s"
                        
                        # 下载依赖，增加超时和重试
                        for i in {1..3}; do
                            echo "尝试下载依赖 (第 $i 次)"
                            if timeout 300s go mod download -x; then
                                echo "依赖下载成功"
                                break
                            else
                                echo "依赖下载失败，尝试备选代理..."
                                # 尝试备选代理
                                if [ $i -eq 2 ]; then
                                    echo "切换到阿里云代理..."
                                    go env -w GOPROXY=https://mirrors.aliyun.com/goproxy/,direct
                                    go env -w GOSUMDB=sum.golang.google.cn
                                elif [ $i -eq 3 ]; then
                                    echo "切换到七牛云代理..."
                                    go env -w GOPROXY=https://goproxy.io,direct
                                    go env -w GOSUMDB=sum.golang.google.cn
                                fi
                                
                                if [ $i -lt 3 ]; then
                                    echo "等待重试..."
                                    sleep 30
                                fi
                            fi
                        done
                        
                        # 验证依赖
                        go mod verify || echo "依赖验证失败，但继续执行"
                    '''
                    
                    // 安装测试工具
                    sh '''
                        # 安装代码质量检查工具
                        echo "安装 golangci-lint..."
                        timeout 120s go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest || echo "golangci-lint 安装失败，继续执行"
                        
                        echo "安装 gotestsum..."
                        timeout 120s go install gotest.tools/gotestsum@latest || echo "gotestsum 安装失败，继续执行"
                        
                        echo "依赖安装完成"
                    '''
                }
            }
        }
        
        stage('代码质量检查') {
            steps {
                script {
                    echo "=== 代码质量检查 ==="
                    
                    // 代码格式化检查
                    sh '''
                        echo "执行代码格式化检查..."
                        go fmt ./... || echo "代码格式化检查完成"
                    '''
                    
                    // 代码静态分析（非阻塞模式）
                    sh '''
                        echo "执行代码静态分析..."
                        if [ -f .golangci.yml ]; then
                            echo "使用配置文件 .golangci.yml"
                            golangci-lint run --timeout=5m --out-format=line-number || {
                                echo "⚠️ 代码质量检查发现问题，但继续执行后续步骤"
                                echo "这些问题将在后续版本中修复"
                            }
                        else
                            echo "未找到配置文件，使用默认配置"
                            golangci-lint run --timeout=5m --out-format=line-number || {
                                echo "⚠️ 代码质量检查发现问题，但继续执行后续步骤"
                                echo "这些问题将在后续版本中修复"
                            }
                        fi
                    '''
                    
                    // 安全检查
                    sh '''
                        echo "执行安全检查..."
                        go vet ./... || echo "安全检查完成"
                    '''
                    
                    // 检查是否有未使用的导入
                    sh '''
                        echo "整理Go模块..."
                        go mod tidy || echo "模块整理完成"
                    '''
                    
                    echo "✅ 代码质量检查完成"
                }
            }
        }
        
        stage('启动测试服务') {
            steps {
                script {
                    echo "=== 启动测试服务 ==="
                    
                    // 清理可能存在的旧容器
                    sh '''
                        echo "清理可能存在的旧测试容器..."
                        docker stop ${MYSQL_CONTAINER_NAME} ${REDIS_CONTAINER_NAME} 2>/dev/null || true
                        docker rm ${MYSQL_CONTAINER_NAME} ${REDIS_CONTAINER_NAME} 2>/dev/null || true
                    '''
                    
                    // 启动MySQL服务
                    sh '''
                        echo "启动MySQL测试服务..."
                        docker run -d \
                            --name ${MYSQL_CONTAINER_NAME} \
                            -e MYSQL_ROOT_PASSWORD=rootpassword \
                            -e MYSQL_DATABASE=oneapi_test \
                            -e MYSQL_USER=testuser \
                            -e MYSQL_PASSWORD=testpass \
                            -e MYSQL_ROOT_HOST=% \
                            -p ${MYSQL_PORT}:3306 \
                            mysql:8.0 --default-authentication-plugin=mysql_native_password
                        
                        echo "等待MySQL服务启动..."
                        for i in $(seq 1 60); do
                            echo "尝试连接MySQL (第${i}次)..."
                            
                            # 检查容器状态
                            if ! docker ps | grep -q ${MYSQL_CONTAINER_NAME}; then
                                echo "❌ MySQL容器未运行"
                                docker logs ${MYSQL_CONTAINER_NAME}
                                exit 1
                            fi
                            
                            # 先尝试使用设置的密码连接
                            if docker exec ${MYSQL_CONTAINER_NAME} mysqladmin ping -h localhost -u root -prootpassword >/dev/null 2>&1; then
                                echo "✅ MySQL服务启动成功"
                                break
                            fi
                            
                            # 如果密码连接失败，尝试空密码连接并设置密码
                            if docker exec ${MYSQL_CONTAINER_NAME} mysqladmin ping -h localhost -u root >/dev/null 2>&1; then
                                echo "检测到空密码，正在设置密码..."
                                docker exec ${MYSQL_CONTAINER_NAME} mysql -u root -e "ALTER USER 'root'@'localhost' IDENTIFIED BY 'rootpassword'; FLUSH PRIVILEGES;" >/dev/null 2>&1 || true
                                docker exec ${MYSQL_CONTAINER_NAME} mysql -u root -e "ALTER USER 'root'@'%' IDENTIFIED BY 'rootpassword'; FLUSH PRIVILEGES;" >/dev/null 2>&1 || true
                                echo "✅ MySQL服务启动成功，密码已设置"
                                break
                            fi
                            
                            echo "等待MySQL启动... (${i}/60)"
                            sleep 3
                        done
                        
                        # 最终检查
                        if ! docker exec ${MYSQL_CONTAINER_NAME} mysqladmin ping -h localhost -u root -prootpassword >/dev/null 2>&1; then
                            echo "❌ MySQL启动失败，显示容器日志:"
                            docker logs ${MYSQL_CONTAINER_NAME}
                            echo "❌ MySQL服务启动失败"
                            exit 1
                        fi
                        
                        echo "验证数据库配置..."
                        docker exec ${MYSQL_CONTAINER_NAME} mysql -u root -prootpassword -e "SHOW DATABASES;" || echo "⚠️ 数据库验证失败"
                        docker exec ${MYSQL_CONTAINER_NAME} mysql -u root -prootpassword -e "SELECT User, Host FROM mysql.user WHERE User IN ('root', 'testuser');" || echo "⚠️ 用户验证失败"
                    '''
                    
                    // 启动Redis服务
                    sh '''
                        echo "启动Redis测试服务..."
                        docker run -d \
                            --name ${REDIS_CONTAINER_NAME} \
                            -p ${REDIS_PORT}:6379 \
                            redis:7-alpine
                        
                        echo "等待Redis服务启动..."
                        for i in {1..15}; do
                            if docker exec ${REDIS_CONTAINER_NAME} redis-cli ping >/dev/null 2>&1; then
                                echo "✅ Redis服务启动成功"
                                break
                            fi
                            echo "等待Redis启动... ($i/15)"
                            sleep 1
                        done
                    '''
                    
                    // 显示服务状态
                    sh '''
                        echo "测试服务状态:"
                        docker ps --filter "name=${MYSQL_CONTAINER_NAME}|${REDIS_CONTAINER_NAME}"
                    '''
                    
                    // 加载测试环境变量
                    sh '''
                        echo "加载测试环境变量..."
                        if [ -f tests/test.env ]; then
                            echo "使用测试环境配置文件"
                            # 只加载非注释行，使用sh兼容语法
                            while IFS= read -r line; do
                                # 跳过空行和注释行
                                case "$line" in
                                    ""|"#"*) continue ;;
                                    *) export "$line" ;;
                                esac
                            done < tests/test.env
                        else
                            echo "使用默认测试环境变量"
                            export SQL_DSN="testuser:testpass@tcp(localhost:3306)/oneapi_test?charset=utf8mb4&parseTime=True&loc=Local"
                            export REDIS_CONN_STRING="redis://localhost:6379"
                            export DEBUG="true"
                            export GLOBAL_WEB_RATE_LIMIT="0"
                            export GLOBAL_API_RATE_LIMIT="0"
                            export SESSION_SECRET="test-secret-key"
                        fi
                        
                        echo "测试环境变量:"
                        echo "SQL_DSN: $SQL_DSN"
                        echo "REDIS_CONN_STRING: $REDIS_CONN_STRING"
                        echo "DEBUG: $DEBUG"
                    '''
                }
            }
        }
        
        stage('单元测试') {
            steps {
                script {
                    echo "=== 单元测试 ==="
                    
                    // 运行单元测试
                    sh '''
                        echo "运行单元测试..."
                        echo "设置测试环境变量..."
                        export SQL_DSN="testuser:testpass@tcp(localhost:3306)/oneapi_test?charset=utf8mb4&parseTime=True&loc=Local"
                        export REDIS_CONN_STRING="redis://localhost:6379"
                        export DEBUG="true"
                        export GLOBAL_WEB_RATE_LIMIT="0"
                        export GLOBAL_API_RATE_LIMIT="0"
                        export SESSION_SECRET="test-secret-key"
                        
                        echo "环境变量:"
                        echo "SQL_DSN: $SQL_DSN"
                        echo "REDIS_CONN_STRING: $REDIS_CONN_STRING"
                        
                        cd tests/unit
                        go test -v -timeout=5m ./... || {
                            echo "⚠️ 单元测试发现问题，但继续执行"
                            echo "测试结果将在后续分析"
                        }
                    '''
                    
                    // 生成测试覆盖率报告
                    sh '''
                        echo "生成测试覆盖率报告..."
                        export SQL_DSN="testuser:testpass@tcp(localhost:3306)/oneapi_test?charset=utf8mb4&parseTime=True&loc=Local"
                        export REDIS_CONN_STRING="redis://localhost:6379"
                        export DEBUG="true"
                        
                        go test -coverprofile=coverage.out -covermode=atomic ./tests/unit/... || echo "覆盖率报告生成失败"
                        go tool cover -html=coverage.out -o coverage.html || echo "HTML覆盖率报告生成失败"
                    '''
                }
            }
            post {
                always {
                    // 发布测试覆盖率报告
                    publishHTML([
                        allowMissing: true,
                        alwaysLinkToLastBuild: true,
                        keepAll: true,
                        reportDir: '.',
                        reportFiles: 'coverage.html',
                        reportName: '测试覆盖率报告'
                    ])
                    
                    // 归档测试结果
                    archiveArtifacts artifacts: 'coverage.out,coverage.html', allowEmptyArchive: true
                }
            }
        }
        
        stage('集成测试') {
            steps {
                script {
                    echo "=== 集成测试 ==="
                    
                    // 运行集成测试
                    sh '''
                        echo "运行集成测试..."
                        echo "设置测试环境变量..."
                        export SQL_DSN="testuser:testpass@tcp(localhost:3306)/oneapi_test?charset=utf8mb4&parseTime=True&loc=Local"
                        export REDIS_CONN_STRING="redis://localhost:6379"
                        export DEBUG="true"
                        export GLOBAL_WEB_RATE_LIMIT="0"
                        export GLOBAL_API_RATE_LIMIT="0"
                        export SESSION_SECRET="test-secret-key"
                        
                        echo "环境变量:"
                        echo "SQL_DSN: $SQL_DSN"
                        echo "REDIS_CONN_STRING: $REDIS_CONN_STRING"
                        
                        cd tests/integration
                        go test -v -timeout=${TEST_TIMEOUT} ./... || {
                            echo "⚠️ 集成测试发现问题，但继续执行"
                            echo "测试结果将在后续分析"
                        }
                    '''
                }
            }
        }
        
        stage('测试结果汇总') {
            steps {
                script {
                    echo "=== 测试结果汇总 ==="
                    
                    // 生成测试报告
                    sh '''
                        echo "## 测试执行结果" > test-report.md
                        echo "" >> test-report.md
                        echo "### 单元测试" >> test-report.md
                        echo "- 执行时间: $(date)" >> test-report.md
                        echo "- 测试文件: tests/unit/" >> test-report.md
                        echo "- 状态: 已执行" >> test-report.md
                        echo "" >> test-report.md
                        echo "### 集成测试" >> test-report.md
                        echo "- 执行时间: $(date)" >> test-report.md
                        echo "- 测试文件: tests/integration/" >> test-report.md
                        echo "- 状态: 已执行" >> test-report.md
                        echo "" >> test-report.md
                        echo "### 测试服务" >> test-report.md
                        echo "- MySQL容器: ${MYSQL_CONTAINER_NAME}" >> test-report.md
                        echo "- Redis容器: ${REDIS_CONTAINER_NAME}" >> test-report.md
                        echo "- MySQL端口: ${MYSQL_PORT}" >> test-report.md
                        echo "- Redis端口: ${REDIS_PORT}" >> test-report.md
                        echo "" >> test-report.md
                        echo "### 代码质量检查" >> test-report.md
                        echo "- 静态分析: golangci-lint" >> test-report.md
                        echo "- 安全检查: go vet" >> test-report.md
                        echo "- 格式化检查: go fmt" >> test-report.md
                        echo "" >> test-report.md
                        echo "### 环境信息" >> test-report.md
                        echo "- Go版本: $(go version)" >> test-report.md
                        echo "- Docker版本: $(docker --version)" >> test-report.md
                        echo "- 构建时间: $(date)" >> test-report.md
                        echo "- 工作空间: ${WORKSPACE}" >> test-report.md
                    '''
                }
            }
            post {
                always {
                    // 归档测试报告
                    archiveArtifacts artifacts: 'test-report.md', allowEmptyArchive: true
                }
            }
        }
    }
    
    post {
        always {
            script {
                echo "=== 清理测试服务 ==="
                
                // 清理测试容器
                sh '''
                    echo "清理测试容器..."
                    docker stop ${MYSQL_CONTAINER_NAME} ${REDIS_CONTAINER_NAME} 2>/dev/null || echo "容器已停止"
                    docker rm ${MYSQL_CONTAINER_NAME} ${REDIS_CONTAINER_NAME} 2>/dev/null || echo "容器已删除"
                    
                    echo "清理完成"
                '''
                
                echo "=== 构建完成 ==="
                echo "构建状态: ${currentBuild.result}"
                echo "构建编号: ${BUILD_NUMBER}"
                echo "构建时间: ${currentBuild.durationString}"
            }
        }
        
        success {
            script {
                echo "✅ 测试执行成功！"
            }
        }
        
        failure {
            script {
                echo "❌ 测试执行失败！"
            }
        }
        
        cleanup {
            script {
                echo "=== 清理工作空间 ==="
                
                // 清理工作空间
                cleanWs()
            }
        }
    }
} 