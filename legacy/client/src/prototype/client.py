#!/usr/bin/python3

import frame
import logging, logging.config
import cProfile
import signal

def signal_handler(signal, frame_py):
	global logging, frame

	print() # clear the signal written to terminal (eg ^C)

	logging.debug("sig:" + str(signal))
	logging.debug("Will now shutdown")
	frame.shutdown()

logging.config.fileConfig("etc/logging.conf")
signal.signal(signal.SIGINT, signal_handler)
frame = frame.frame()
frame.main_loop()
