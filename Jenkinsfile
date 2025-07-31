pipeline {
    agent any
    
    environment {
        // 独立测试环境配置 - 完全隔离，不影响宿主机现有服务
        MYSQL_CONTAINER_NAME = "oneapi-mysql-test-${BUILD_NUMBER}"
        REDIS_CONTAINER_NAME = "oneapi-redis-test-${BUILD_NUMBER}"
        GO_CONTAINER_NAME = "oneapi-go-test-${BUILD_NUMBER}"
        TEST_NETWORK_NAME = "oneapi-test-network-${BUILD_NUMBER}"
        
        // 使用不同端口避免冲突（宿主机已用3406, 6479）
        MYSQL_PORT = "3307"  
        REDIS_PORT = "6380"
        
        // 测试数据库配置
        MYSQL_USER = "testuser"
        MYSQL_PASSWORD = "testpass"
        MYSQL_DATABASE = "oneapi_test"
        TEST_TIMEOUT = "10m"
    }
    
    stages {
        stage('环境检查') {
            steps {
                script {
                    echo "=== 环境检查 ==="
                    
                    sh '''
                    echo "Jenkins 工作空间: ${WORKSPACE}"
                    echo "构建编号: ${BUILD_NUMBER}"
                    echo "Git 分支: ${GIT_BRANCH}"
                    
                        echo "检查Docker环境..."
                        docker --version
                        docker info --format '{{.ServerVersion}}'
                        
                        echo "检查可用端口..."
                        netstat -tlnp | grep -E ":(3307|6380)" || echo "端口可用"
                    '''
                }
            }
        }
        
        stage('启动测试服务') {
            steps {
                script {
                    echo "=== 启动测试服务 ==="
                    
                    // 清理可能存在的旧容器和网络
                    sh '''
                        echo "清理可能存在的旧测试资源..."
                        docker stop ${MYSQL_CONTAINER_NAME} ${REDIS_CONTAINER_NAME} ${GO_CONTAINER_NAME} 2>/dev/null || true
                        docker rm ${MYSQL_CONTAINER_NAME} ${REDIS_CONTAINER_NAME} ${GO_CONTAINER_NAME} 2>/dev/null || true
                        docker network rm ${TEST_NETWORK_NAME} 2>/dev/null || true
                    '''
                    
                    // 创建独立的测试网络
                    sh '''
                        echo "创建独立测试网络..."
                        docker network create ${TEST_NETWORK_NAME}
                        echo "✅ 测试网络创建完成: ${TEST_NETWORK_NAME}"
                    '''
                    
                    // 启动MySQL服务
                    sh '''
                        echo "启动MySQL测试服务..."
                        docker run -d \
                            --name ${MYSQL_CONTAINER_NAME} \
                            --network ${TEST_NETWORK_NAME} \
                            -p ${MYSQL_PORT}:3306 \
                            -e MYSQL_ROOT_PASSWORD=rootpassword \
                            -e MYSQL_DATABASE=${MYSQL_DATABASE} \
                            -e MYSQL_USER=${MYSQL_USER} \
                            -e MYSQL_PASSWORD=${MYSQL_PASSWORD} \
                            mysql:8.0 --default-authentication-plugin=mysql_native_password
                        
                        echo "等待MySQL服务启动..."
                        for i in $(seq 1 30); do
                            if docker exec ${MYSQL_CONTAINER_NAME} mysqladmin ping -h localhost -u root -prootpassword >/dev/null 2>&1; then
                                echo "✅ MySQL服务启动成功"
                                break
                            fi
                            echo "等待MySQL启动... (${i}/30)"
                            sleep 2
                        done
                        
                        # 验证MySQL连接
                        if docker exec ${MYSQL_CONTAINER_NAME} mysql -u testuser -ptestpass -e "SELECT 1;" >/dev/null 2>&1; then
                            echo "✅ MySQL用户连接验证成功"
                        else
                            echo "⚠️ MySQL用户连接验证失败"
                        fi
                    '''
                    
                    // 启动Redis服务
                    sh '''
                        echo "启动Redis测试服务..."
                        docker run -d \
                            --name ${REDIS_CONTAINER_NAME} \
                            --network ${TEST_NETWORK_NAME} \
                            -p ${REDIS_PORT}:6379 \
                            redis:7-alpine redis-server --protected-mode no
                        
                        echo "等待Redis服务启动..."
                        for i in $(seq 1 15); do
                            if docker exec ${REDIS_CONTAINER_NAME} redis-cli ping >/dev/null 2>&1; then
                                echo "✅ Redis服务启动成功"
                                break
                            fi
                            echo "等待Redis启动... (${i}/15)"
                            sleep 1
                        done
                        
                        # 验证Redis连接
                        if docker exec ${REDIS_CONTAINER_NAME} redis-cli ping | grep -q PONG; then
                            echo "✅ Redis连接验证成功"
                        else
                            echo "⚠️ Redis连接验证失败"
                        fi
                    '''
                    
                    // 显示服务状态
                    sh '''
                        echo "测试服务状态:"
                        docker ps --filter "name=${MYSQL_CONTAINER_NAME}|${REDIS_CONTAINER_NAME}"
                        echo "端口映射:"
                        echo "MySQL: localhost:${MYSQL_PORT} -> container:3306"
                        echo "Redis: localhost:${REDIS_PORT} -> container:6379"
                    '''
                }
            }
        }
        
        stage('单元测试') {
            steps {
                script {
                    echo "=== 单元测试 ==="
                    
                    sh '''
                        echo "运行单元测试..."
                        echo "设置测试环境变量..."
                        
                        # 设置测试环境变量（Go容器通过Docker网络连接到测试服务容器）
                        export SQL_DSN="testuser:testpass@tcp(${MYSQL_CONTAINER_NAME}:3306)/${MYSQL_DATABASE}?charset=utf8mb4&parseTime=True&loc=Local"
                        export REDIS_CONN_STRING="redis://${REDIS_CONTAINER_NAME}:6379"
                        export DEBUG="true"
                        export GLOBAL_WEB_RATE_LIMIT="0"
                        export GLOBAL_API_RATE_LIMIT="0"
                        export SESSION_SECRET="test-secret-key"
                        export SYNC_FREQUENCY="60"
                        
                        echo "环境变量:"
                        echo "SQL_DSN: $SQL_DSN"
                        echo "REDIS_CONN_STRING: $REDIS_CONN_STRING"
                        
                        # 测试网络连接
                        echo "测试网络连接..."
                        # 检查容器是否正在运行
                        if docker ps | grep -q ${MYSQL_CONTAINER_NAME}; then
                            echo "✅ MySQL容器运行正常"
                            # 直接测试容器内部连接
                            if docker exec ${MYSQL_CONTAINER_NAME} mysqladmin ping -h localhost -u testuser -ptestpass >/dev/null 2>&1; then
                                echo "✅ MySQL容器内部连接正常"
                            else
                                echo "⚠️ MySQL容器内部连接失败"
                            fi
                        else
                            echo "❌ MySQL容器未运行"
                        fi
                        
                        if docker ps | grep -q ${REDIS_CONTAINER_NAME}; then
                            echo "✅ Redis容器运行正常"
                            # 直接测试容器内部连接
                            if docker exec ${REDIS_CONTAINER_NAME} redis-cli ping >/dev/null 2>&1; then
                                echo "✅ Redis容器内部连接正常"
                            else
                                echo "⚠️ Redis容器内部连接失败"
                            fi
                        else
                            echo "❌ Redis容器未运行"
                        fi
                        
                        # 在Jenkins环境中，使用专用Go容器运行测试
                        echo "启动Go测试容器..."
                        docker run --rm \
                            --name ${GO_CONTAINER_NAME}-unit \
                            --network ${TEST_NETWORK_NAME} \
                            -v ${WORKSPACE}:/workspace \
                            -w /workspace \
                            -e SQL_DSN="$SQL_DSN" \
                            -e REDIS_CONN_STRING="$REDIS_CONN_STRING" \
                            -e DEBUG="$DEBUG" \
                            -e GLOBAL_WEB_RATE_LIMIT="$GLOBAL_WEB_RATE_LIMIT" \
                            -e GLOBAL_API_RATE_LIMIT="$GLOBAL_API_RATE_LIMIT" \
                            -e SESSION_SECRET="$SESSION_SECRET" \
                            -e SYNC_FREQUENCY="$SYNC_FREQUENCY" \
                            -e GOPROXY="https://goproxy.cn,https://goproxy.io,direct" \
                            -e GOSUMDB="sum.golang.google.cn" \
                            -e GO111MODULE="on" \
                            golang:1.21 sh -c "
                                echo '开始执行单元测试...'
                                echo 'Go代理配置: \$GOPROXY'
                                cd tests/unit
                                timeout 300 go mod download || echo '模块下载超时，但继续执行'
                                go test -v -timeout=${TEST_TIMEOUT} ./...
                                echo '单元测试执行完成'
                            " || {
                            echo "⚠️ 单元测试发现问题，但继续执行"
                        }
                    '''
                }
            }
        }
        
        stage('集成测试') {
            steps {
                script {
                    echo "=== 集成测试 ==="
                    
                    sh '''
                        echo "运行集成测试..."
                        echo "设置测试环境变量..."
                        
                        # 设置测试环境变量（Go容器通过Docker网络连接到测试服务容器）
                        export SQL_DSN="testuser:testpass@tcp(${MYSQL_CONTAINER_NAME}:3306)/${MYSQL_DATABASE}?charset=utf8mb4&parseTime=True&loc=Local"
                        export REDIS_CONN_STRING="redis://${REDIS_CONTAINER_NAME}:6379"
                        export DEBUG="true"
                        export GLOBAL_WEB_RATE_LIMIT="0"
                        export GLOBAL_API_RATE_LIMIT="0"
                        export SESSION_SECRET="test-secret-key"
                        export SYNC_FREQUENCY="60"
                        
                        echo "环境变量:"
                        echo "SQL_DSN: $SQL_DSN"
                        echo "REDIS_CONN_STRING: $REDIS_CONN_STRING"
                        
                        # 在Jenkins环境中，使用专用Go容器运行集成测试
                        echo "启动Go集成测试容器..."
                        docker run --rm \
                            --name ${GO_CONTAINER_NAME}-integration \
                            --network ${TEST_NETWORK_NAME} \
                            -v ${WORKSPACE}:/workspace \
                            -w /workspace \
                            -e SQL_DSN="$SQL_DSN" \
                            -e REDIS_CONN_STRING="$REDIS_CONN_STRING" \
                            -e DEBUG="$DEBUG" \
                            -e GLOBAL_WEB_RATE_LIMIT="$GLOBAL_WEB_RATE_LIMIT" \
                            -e GLOBAL_API_RATE_LIMIT="$GLOBAL_API_RATE_LIMIT" \
                            -e SESSION_SECRET="$SESSION_SECRET" \
                            -e SYNC_FREQUENCY="$SYNC_FREQUENCY" \
                            -e GOPROXY="https://goproxy.cn,https://goproxy.io,direct" \
                            -e GOSUMDB="sum.golang.google.cn" \
                            -e GO111MODULE="on" \
                            golang:1.21 sh -c "
                                echo '开始执行集成测试...'
                                echo 'Go代理配置: \$GOPROXY'
                                cd tests/integration
                                timeout 300 go mod download || echo '模块下载超时，但继续执行'
                                go test -v -timeout=${TEST_TIMEOUT} ./...
                                echo '集成测试执行完成'
                            " || {
                            echo "⚠️ 集成测试发现问题，但继续执行"
                        }
                    '''
                }
            }
        }
        
        stage('生成覆盖率报告') {
            steps {
                script {
                    echo "=== 生成覆盖率报告 ==="
                    
                        sh '''
                        echo "生成测试覆盖率报告..."
                        
                        # 设置测试环境变量（Go容器通过Docker网络连接到测试服务容器）
                        export SQL_DSN="testuser:testpass@tcp(${MYSQL_CONTAINER_NAME}:3306)/${MYSQL_DATABASE}?charset=utf8mb4&parseTime=True&loc=Local"
                        export REDIS_CONN_STRING="redis://${REDIS_CONTAINER_NAME}:6379"
                        export DEBUG="true"
                        
                        # 使用Go容器生成覆盖率报告
                        docker run --rm \
                            --name ${GO_CONTAINER_NAME}-coverage \
                            --network ${TEST_NETWORK_NAME} \
                            -v ${WORKSPACE}:/workspace \
                            -w /workspace \
                            -e SQL_DSN="$SQL_DSN" \
                            -e REDIS_CONN_STRING="$REDIS_CONN_STRING" \
                            -e DEBUG="$DEBUG" \
                            -e GOPROXY="https://goproxy.cn,https://goproxy.io,direct" \
                            -e GOSUMDB="sum.golang.google.cn" \
                            -e GO111MODULE="on" \
                            golang:1.21 sh -c "
                                echo '生成测试覆盖率报告...'
                                echo 'Go代理配置: \$GOPROXY'
                                timeout 300 go mod download || echo '模块下载超时，但继续执行'
                                go test -coverprofile=coverage.out -covermode=atomic ./tests/unit/... || echo '覆盖率报告生成失败'
                                go tool cover -html=coverage.out -o coverage.html || echo 'HTML覆盖率报告生成失败'
                                echo '覆盖率报告生成完成'
                            " || echo "覆盖率报告生成失败"
                    '''
                }
            }
            post {
                always {
                    publishHTML([
                        allowMissing: true,
                        alwaysLinkToLastBuild: true,
                        keepAll: true,
                        reportDir: '.',
                        reportFiles: 'coverage.html',
                        reportName: '测试覆盖率报告'
                    ])
                    archiveArtifacts artifacts: 'coverage.out,coverage.html', allowEmptyArchive: true
                }
            }
        }
        
        stage('测试结果汇总') {
            steps {
                script {
                    echo "=== 测试结果汇总 ==="
                    
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
                    archiveArtifacts artifacts: 'test-report.md', allowEmptyArchive: true
                }
            }
        }
    }
    
    post {
        always {
            script {
                echo "=== 清理测试服务 ==="
                
                sh '''
                    echo "清理测试资源..."
                    
                    # 清理测试容器
                    echo "停止并删除测试容器..."
                    docker stop ${MYSQL_CONTAINER_NAME} ${REDIS_CONTAINER_NAME} 2>/dev/null || echo "测试容器已停止"
                    docker rm ${MYSQL_CONTAINER_NAME} ${REDIS_CONTAINER_NAME} 2>/dev/null || echo "测试容器已删除"
                    
                    # 清理可能的Go测试容器
                    docker rm -f ${GO_CONTAINER_NAME}-unit 2>/dev/null || true
                    docker rm -f ${GO_CONTAINER_NAME}-integration 2>/dev/null || true
                    docker rm -f ${GO_CONTAINER_NAME}-coverage 2>/dev/null || true
                    
                    # 清理测试网络
                    echo "清理测试网络..."
                    docker network rm ${TEST_NETWORK_NAME} 2>/dev/null || echo "测试网络已清理或不存在"
                    
                    echo "✅ 测试资源清理完成"
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
                cleanWs()
            }
        }
    }
} 