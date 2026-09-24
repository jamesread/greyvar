import pygame
import pygame.time
import logging
import time

class renderer:		
	MESSAGE_DISPLAY_TIME = 4; 
	COLOR_TRANSPARENT = (66, 66, 66)
	messages = {} 
	players = []
	textures = {}

	def __init__(self, screen, m): 
		self.screen = screen;
		self.clock = pygame.time.Clock() 
#		self.screen.set_clip((10, 10, 100, 100))
		self.m = m;
		self.screen.fill((0, 100, 175));
		pygame.display.flip();
  
		self.defaultTexture = self.loadTex("tiles/construct.png");
 		
		self.serverHostname = None
		 
		self.surfVoid = pygame.Surface(self.screen.get_size()).convert();
		self.initSurfCharacters();
		self.initSurfHud();
		self.initSurfgrid();
		self.initFonts();

	def toggleSurf(self, surf):
		surf.set_alpha((100 if surf.get_alpha() == 0 else 0))

	def scroll(self, x, y):
		self.surfgridX += x * 48;
		self.surfgridY += y * 48;
		self.surfCharactersX += x * 48;
		self.surfCharactersY += y * 48;
		pygame.display.flip()
		
	def createTextAndPosition(self, text, color, x, y):
		txt = self.fontDefault.render(text, False, color);	
		pos = txt.get_rect()
		pos.x = x;
		pos.y = y;
		
		return [txt, pos]

	def initFonts(self):
		self.fontDefault = pygame.font.Font("res/ttf/DejaVuSansMono.ttf", 24);
		
	def initSurfCharacters(self):
		self.surfCharacters = pygame.Surface(self.screen.get_size()).convert()
		self.surfCharactersX = 0;
		self.surfCharactersY = 0;   
		self.surfCharacters.set_colorkey((10,10,10)) 

	def initSurfgrid(self):
		width, height = self.screen.get_size()

		self.surfgrid = pygame.Surface((width, height)).convert()
		self.surfgridX = 0;
		self.surfgridY = 0;

	def initSurfHud(self):
		width, height = self.screen.get_size();
 
		self.surfHud = pygame.Surface((width, height)).convert()
		self.surfHud.set_colorkey(self.COLOR_TRANSPARENT);
 
	def addPlayer(self, player): 
		self.players.append(player); 
		
	def addMessage(self, message):  
		self.messages[time.time()] = message;

	def draw(self):   
		self.clock.tick();
		
		self.drawVoid();
		self.drawgrid()
		self.drawCharacters();  
		self.drawHud();
		
		pygame.display.flip(); 
  
	def drawHud(self): 
		if self.surfHud.get_alpha() == 0: return;
 		             
		self.surfHud.fill(self.COLOR_TRANSPARENT)        
		pygame.draw.rect(self.surfHud, (22,22,22), (0, 0, 768, 32))

		txtBanner = self.fontDefault.render("Greyvar v0.0", False, (120, 120, 120));
		pos = txtBanner.get_rect();
		pos.x += 4;
		pos.y += 4;
		self.surfHud.blit(txtBanner, pos)

		txtHostname = self.fontDefault.render("Server: " + (self.serverHostname if not self.serverHostname == None else "NOT CONNECTED"), False, (120, 120, 120))
		pos = txtHostname.get_rect();
		pos.x = 320;
		pos.y = 4;
		self.surfHud.blit(txtHostname, pos)
		   
		txt, pos = self.createTextAndPosition("FPS: " + str(int(self.clock.get_fps())), (50, 150, 50), 640, 4)
		self.surfHud.blit(txt, pos);
		 
		now = time.time()
		renderedMessages = 0    
		for timestamp in self.messages.copy(): 
			if (timestamp + self.MESSAGE_DISPLAY_TIME < now): 
				logging.debug("Deleting old message: " + self.messages[timestamp]);
				del self.messages[timestamp]
			else:  
				txtMessage = self.fontDefault.render(self.messages[timestamp], False, (150, 50, 50))
				pos = txtMessage.get_rect()
				pos.x = 20    
				pos.y = 600 - ((len(self.messages) - renderedMessages) * pos.h) 
				self.surfHud.blit(txtMessage, pos) 
				 
				renderedMessages += 1
	 
#		textpos.centerx = self.surfHud.get_rect().centerx
#		self.surfHud.set_colorkey((0,0,0)) 
		self.screen.blit(self.surfHud, (0, 0))
 
	def drawCharacters(self):  
		self.surfCharacters.fill((10,10,10));
		for player in self.players: 
			if (time.time() - player.lastMoved) > .5:
				self.surfCharacters.blit(self.getTex("entities/player" + player.tex + ".png"), (player.x * 48, player.y * 48))
			else:
				if player.lastX < player.x:
					self.surfCharacters.blit(self.getTex("entities/player" + player.tex + "Right.png"), (player.x * 48, player.y * 48))
				elif player.lastX > player.x:
					self.surfCharacters.blit(self.getTex("entities/player" + player.tex + "Left.png"), (player.x * 48, player.y * 48))
				elif player.lastY < player.y:
					self.surfCharacters.blit(self.getTex("entities/player" + player.tex + ".png"), (player.x * 48, player.y * 48))
				elif player.lastY > player.y:
					self.surfCharacters.blit(self.getTex("entities/player" + player.tex + "Up.png"), (player.x * 48, player.y * 48))
				else:
					self.surfCharacters.blit(self.getTex("entities/player" + player.tex + ".png"), (player.x * 48, player.y * 48))
			 
		self.screen.blit(self.surfCharacters, (self.surfCharactersX, self.surfCharactersY));

	def loadTex(self, filename): 
		img = pygame.image.load("res/img/textures/" + filename);
		img = pygame.transform.scale(img, (48, 48));

		self.textures[filename] = img;

		return img;
	
	def drawVoid(self): 
		self.surfVoid.fill((100, 0, 0)) 
		self.screen.blit(self.surfVoid, (0,0));
 
	def drawgrid(self):
		for row in range(0, 16):
			for col in range(0, 16):
				tile = self.m.getTile(row, col);

				self.drawTile(tile, row, col);
 				
		self.screen.blit(self.surfgrid, (self.surfgridX, self.surfgridY));

	def getTex(self, tex):
		if not tex in self.textures and not self.loadTex(tex):
			logging.warn("Warning, texture not loaded:" + tex);
			return self.defaultTexture;

		return self.textures[tex];

	def drawTile(self, tile, row, col):
		width = 48;
		height = 48;
		tex = self.getTex("tiles/" + tile.tex) 
		tex = pygame.transform.rotate(tex, -tile.rot)
		tex = pygame.transform.flip(tex, tile.flipV, tile.flipH)
#		tex = pygame.transform.flip(tex, tile.flipV, tile.flipH)
#		tex = pygame.transform.flip(tex, True, True)

		
		self.surfgrid.blit(tex, (row * width, col * height));
