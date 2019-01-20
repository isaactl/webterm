pipeline {
  agent {
    docker {
      image 'golang:1.10-alpine'
    }

  }
  stages {
    stage('error') {
      steps {
        sh 'ls -alh'
      }
    }
  }
}