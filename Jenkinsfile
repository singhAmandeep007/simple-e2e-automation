pipeline {
  agent any

  environment {
    NVM_DIR         = "${env.HOME}/.nvm"
    COMPOSE_PROJECT = "poc-e2e-${env.BUILD_NUMBER}"
    CP_PORT         = "4000"
    UI_PORT         = "5173"
  }

  options {
    timeout(time: 30, unit: 'MINUTES')
    disableConcurrentBuilds()
    buildDiscarder(logRotator(numToKeepStr: '10'))
  }

  stages {

    stage('Checkout') {
      steps { checkout scm }
    }

    // ── Docker-first CI (recommended) ────────────────────────────────────────
    stage('Build Docker Images') {
      when { not { branch 'skip-docker' } }
      steps {
        sh '''
          echo "Building all Docker images..."
          docker compose -p ${COMPOSE_PROJECT} build --parallel
          # Also build e2e image (separate profile build)
          docker compose -p ${COMPOSE_PROJECT} --profile e2e build e2e
          echo "Docker images built successfully"
        '''
      }
    }

    stage('E2E Tests (Docker)') {
      steps {
        sh '''
          echo "Running full e2e pipeline in Docker..."
          docker compose -p ${COMPOSE_PROJECT} \
            --profile e2e \
            up \
            --exit-code-from e2e \
            --abort-on-container-exit
        '''
      }
      post {
        always {
          // Copy test results out of Docker
          sh '''
            docker compose -p ${COMPOSE_PROJECT} --profile e2e \
              cp e2e:/e2e/cypress/results ./cypress-results 2>/dev/null || true
            docker compose -p ${COMPOSE_PROJECT} --profile e2e \
              cp e2e:/e2e/cypress/videos ./cypress-videos 2>/dev/null || true
            docker compose -p ${COMPOSE_PROJECT} --profile e2e \
              cp e2e:/e2e/cypress/screenshots ./cypress-screenshots 2>/dev/null || true
          '''
          junit allowEmptyResults: true, testResults: 'cypress-results/**/*.xml'
          archiveArtifacts artifacts: 'cypress-videos/**,cypress-screenshots/**',
                           allowEmptyArchive: true
        }
        cleanup {
          sh "docker compose -p ${COMPOSE_PROJECT} --profile e2e down -v --remove-orphans || true"
        }
      }
    }

    // ── Fallback: Local (non-Docker) CI ──────────────────────────────────────
    // Enable by setting JENKINS_LOCAL_CI=true in build parameters
    stage('Setup Go + Node (Local CI)') {
      when { environment name: 'JENKINS_LOCAL_CI', value: 'true' }
      steps {
        sh '''
          echo "Building Go binaries..."
          mkdir -p bin
          cd agent && go build -o ../bin/go-agent .
          cd ../control-plane && go build -o ../bin/control-plane .
          echo "Installing Node dependencies..."
          source ${NVM_DIR}/nvm.sh && nvm use
          npm install
          npm install --prefix web-ui
          npm install --prefix e2e
        '''
      }
    }

    stage('E2E Tests (Local CI)') {
      when { environment name: 'JENKINS_LOCAL_CI', value: 'true' }
      steps {
        sh '''
          source ${NVM_DIR}/nvm.sh && nvm use
          mkdir -p control-plane/data
          (cd control-plane && ../bin/control-plane > /tmp/cp-${BUILD_NUMBER}.log 2>&1) &
          echo $! > /tmp/cp-${BUILD_NUMBER}.pid
          sleep 3
          npm run dev --prefix web-ui > /tmp/ui-${BUILD_NUMBER}.log 2>&1 &
          echo $! > /tmp/ui-${BUILD_NUMBER}.pid
          sleep 5
          cd e2e && npm run cy:ci || EXIT=$?
          kill $(cat /tmp/cp-${BUILD_NUMBER}.pid) 2>/dev/null || true
          kill $(cat /tmp/ui-${BUILD_NUMBER}.pid) 2>/dev/null || true
          exit ${EXIT:-0}
        '''
      }
      post {
        always {
          junit allowEmptyResults: true, testResults: 'e2e/cypress/results/**/*.xml'
          archiveArtifacts artifacts: 'e2e/cypress/videos/**,e2e/cypress/screenshots/**',
                           allowEmptyArchive: true
        }
      }
    }

  }

  post {
    failure {
      echo 'Pipeline failed. Check archived artifacts for screenshots/videos.'
    }
    always {
      cleanWs()
    }
  }
}
