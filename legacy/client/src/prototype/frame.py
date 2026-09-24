import pygame, pygame.event
import renderer
import input_handler
from server_interface import server_interface
import grid
import logging
import clgame
import sys

class frame:
	def __init__(self):
		logging.debug("client init");
		#pygame.init();
		pygame.font.init()
		
		self.running = True

		self.m = grid.grid()
		self.start_frame();
		self.start_renderer();  

		self.clgame = clgame.clgame(self.m);
		self.clgame.frame = self
		self.si = server_interface(self, self.r, self.clgame);
		self.ih = input_handler.input_handler(self.si, self.clgame, self.r);

	def start_frame(self):
		favicon = pygame.image.load("res/img/logo.png");

		pygame.display.set_icon(favicon);
		pygame.display.set_caption("Greyvar");

		self.screen = pygame.display.set_mode((48 * 16, 48 * 16))

		logging.debug("started frame");

	def start_renderer(self):
		logging.debug("Start renderer");

		self.r = renderer.renderer(self.screen, self.m);	
		self.r.addMessage("Greyvar");
		self.r.addMessage("---");
		self.r.addMessage("F10: Connect");

		logging.debug("Finished loading renderer")
		
	def main_loop(self): 
		while self.running:  
			for event in pygame.event.get(): 
				if event.type == pygame.QUIT: 
					self.shutdown()
					return; 
				elif event.type == pygame.KEYDOWN:
					self.ih.handleKeypress(event); 
					
			pygame.time.delay(50);
			self.r.draw();
   
		self.running = False

	def shutdown(self):
		logging.info("frame Shutdown()")

		logging.info("Stopping main loop")
		self.running = False

		logging.info("Waiting for server interface to disconnect")
		self.si.disconnect();

		logging.info("Waiting for PyGame to quit")
		pygame.quit() 

		logging.info("Exiting!")
		sys.exit(0);

