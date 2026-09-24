import time

class player:
	def __init__(self, id, nickname):
		self.id = id;
		self.nickname = nickname;
		self.tex = "playerBlue.png"
		self.x = 5
		self.y = 5
		self.lastX = 5
		self.lastY = 5
		self.lastMoved = time.time()
