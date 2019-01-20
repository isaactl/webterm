pipeline {
  agent {
    docker {
      image 'golang:1.10-alpine'
    }

  }
  stages {
    stage('build') {
      steps {
        sh 'ls -alh'
        sh 'go build'
      }
    }
  }
}