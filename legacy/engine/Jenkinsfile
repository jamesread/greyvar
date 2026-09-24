stage("Compile") {
	node {
		checkout scm
		sh "cmake ."
		sh "make"
		sh "./package.sh"
		archive "pkg/*.zip"
	}
}
