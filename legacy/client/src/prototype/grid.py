import tile

class grid:
	def __init__(self):
		self.tiles = {}

		for row in range(0, 16):
			self.tiles[row] = {}

			for col in range(0, 16):
				self.tiles[row][col] = None
 
	def getTile(self, x, y):
		if self.tiles[x][y] == None:
			t = tile.tile();
			t.tex = "grass.png"
			return t
		else:
			return self.tiles[x][y]
		

	def setTile(self, row, col, tile):
		self.tiles[row][col] = tile;

