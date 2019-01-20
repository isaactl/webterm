pipeline {
  agent {
    docker {
      image 'golang:1.10-alpine'
      args 'build env'
    }

  }
  stages {
    stage('') {
      steps {
        sh 'ls -alh'
      }
    }
  }
}