#pragma once

#ifndef H_FOO
#define H_FOO

#include <pthread.h>
#include <iostream>

using std::cout;
using std::endl;
using std::string;

void *runUpdateCheck(void *threadId) {
	cout << "Running Check on thread " << (long)threadId << endl;	
	pthread_exit(NULL);
}

#endif
