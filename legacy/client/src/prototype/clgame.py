import player

class clgame:
	frame = None

	def __init__(self, m):
		self.players = {}
		self.m = m

	def addPlayer(self, p, playerId):
		assert isinstance(p, player.player)
		self.players[playerId] = p

	def getPlayer(self, playerId):
		return self.players[playerId]
