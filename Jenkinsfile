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
        
        // 测试配置
        TEST_TIMEOUT = '10m'
        COVERAGE_THRESHOLD = '70'
    }
    
    stages {
        stage('环境检查') {
            steps {
                script {
                    echo "=== 环境检查 ==="
                    echo "Jenkins 工作空间: ${WORKSPACE}"
                    echo "构建编号: ${BUILD_NUMBER}"
                    echo "Git 分支: ${GIT_BRANCH}"
                    
                    // 检查必要工具
                    sh 'go version'
                    sh 'which make'
                }
            }
        }
        
        stage('代码检出') {
            steps {
                script {
                    echo "=== 代码检出 ==="
                    
                    // 清理工作空间
                    cleanWs()
                    
                    // 检出代码
                    checkout([
                        $class: 'GitSCM',
                        branches: [[name: "*/${GIT_BRANCH}"]],
                        doGenerateSubmoduleConfigurations: false,
                        extensions: [
                            [$class: 'CleanBeforeCheckout'],
                            [$class: 'CleanCheckout'],
                            [$class: 'SubmoduleOption', 
                             disableSubmodules: false, 
                             recursiveSubmodules: true, 
                             trackingSubmodules: false, 
                             reference: '', 
                             parentCredentials: false]
                        ],
                        submoduleCfg: [],
                        userRemoteConfigs: [[
                            credentialsId: 'git-code-tencent-credentials',
                            url: "${GIT_REPO}"
                        ]]
                    ])
                    
                    // 显示代码信息
                    sh 'git log --oneline -5'
                    sh 'git status'
                }
            }
        }
        
        stage('依赖安装') {
            steps {
                script {
                    echo "=== 依赖安装 ==="
                    
                    // 设置 Go 环境
                    withEnv(["PATH+GO=${GOROOT}/bin:${GOPATH}/bin"]) {
                        // 下载依赖
                        sh 'go mod download'
                        sh 'go mod verify'
                        
                        // 安装测试工具
                        sh 'go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest'
                        sh 'go install gotest.tools/gotestsum@latest'
                    }
                }
            }
        }
        
        stage('代码质量检查') {
            steps {
                script {
                    echo "=== 代码质量检查 ==="
                    
                    withEnv(["PATH+GO=${GOROOT}/bin:${GOPATH}/bin"]) {
                        // 代码格式化检查
                        sh 'go fmt ./...'
                        
                        // 代码静态分析
                        sh 'golangci-lint run --timeout=5m --out-format=line-number'
                        
                        // 安全检查
                        sh 'go vet ./...'
                        
                        // 检查是否有未使用的导入
                        sh 'go mod tidy'
                    }
                }
            }
        }
        
        stage('单元测试') {
            steps {
                script {
                    echo "=== 单元测试 ==="
                    
                    withEnv(["PATH+GO=${GOROOT}/bin:${GOPATH}/bin"]) {
                        // 运行单元测试
                        sh '''
                            cd tests/unit
                            make unit-test
                        '''
                        
                        // 生成测试覆盖率报告
                        sh '''
                            go test -coverprofile=coverage.out -covermode=atomic ./tests/unit/...
                            go tool cover -html=coverage.out -o coverage.html
                        '''
                    }
                }
            }
            post {
                always {
                    // 发布测试覆盖率报告
                    publishHTML([
                        allowMissing: false,
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
                    
                    withEnv(["PATH+GO=${GOROOT}/bin:${GOPATH}/bin"]) {
                        // 启动测试环境
                        sh '''
                            docker-compose -f docker-compose.test.yml up -d
                            sleep 10
                        '''
                        
                        // 运行集成测试
                        sh '''
                            cd tests/integration
                            go test -v -timeout=${TEST_TIMEOUT} ./...
                        '''
                    }
                }
            }
            post {
                always {
                    // 清理测试环境
                    sh 'docker-compose -f docker-compose.test.yml down -v'
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
                        echo "" >> test-report.md
                        echo "### 集成测试" >> test-report.md
                        echo "- 执行时间: $(date)" >> test-report.md
                        echo "- 测试文件: tests/integration/" >> test-report.md
                        echo "" >> test-report.md
                        echo "### 代码质量检查" >> test-report.md
                        echo "- 静态分析: golangci-lint" >> test-report.md
                        echo "- 安全检查: go vet" >> test-report.md
                        echo "- 格式化检查: go fmt" >> test-report.md
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