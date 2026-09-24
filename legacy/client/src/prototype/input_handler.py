import pygame
from pygame.locals import *
import logging
 
class input_handler:
	def __init__(self, si, clgame, renderer):
		self.si = si;
		self.clgame = clgame
		self.renderer = renderer
		      
		pygame.key.set_repeat(72, 250)

	def handleKeypress(self, event):
		connected = self.si.connected

		if   connected and event.key == K_DOWN: 
			self.si.send_move(0, +1); 
		elif connected and event.key == K_UP:
			self.si.send_move(0, -1);
		elif connected and event.key == K_LEFT:
			self.si.send_move(-1, 0);
		elif connected and event.key == K_RIGHT:
			self.si.send_move(+1, 0);
		elif not connected and event.key == K_F10: 
			self.si.connect()
			self.si.send_helo()
		elif connected and event.key == K_F9: 
			self.si.send_halt()
		elif event.key == K_F8: 
			self.renderer.toggleSurf(self.renderer.surfHud)
		elif event.key == K_F7: 
			self.renderer.toggleSurf(self.renderer.surfgrid)
		elif event.key == K_w: 
			self.renderer.scroll(0, 1);
		elif event.key == K_a:
			self.renderer.scroll(1, 0);
		elif event.key == K_s:
			self.renderer.scroll(0, -1);
		elif event.key == K_d: 
			self.renderer.scroll(-1, 0);
		elif event.key == K_ESCAPE:
			self.clgame.frame.shutdown();
		else:  
			logging.debug("Unknown key: " + str(event.key))
