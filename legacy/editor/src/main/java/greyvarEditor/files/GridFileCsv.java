package greyvarEditor.files;

import greyvarEditor.GameResources;
import greyvarEditor.TextureCache;
import greyvarEditor.Tile;
import greyvarEditor.Tile.TeleportDirection;
import greyvarEditor.dataModel.EntityInstance;
import greyvarEditor.ui.windows.editors.grid.panels.Texture;
import greyvarEditor.utils.CsvFile;

import java.io.BufferedReader;
import java.io.File;
import java.io.FileNotFoundException;
import java.io.FileReader;
import java.io.FileWriter;
import java.io.IOException;
import java.util.Arrays;
import java.util.Calendar;
import java.util.Vector;

import jwrCommonsJava.Logger;

class GridFileCsv extends CsvFile implements GridFile {
	private Grid<Tile> tiles;
	
	private int gridWidth;
	private int gridHeight;

	public GridFileCsv(File f) {
		super(f);

		this.initMap();
		this.load();
	}
	
	@Override
	public File getFile() {
		return this.f; 
	}

	public GridFileCsv(String text) {
		this(new File(GameResources.dirServerData + "/grids/" + text));
	}

 
	public Grid<Tile> getTileList() {
		return this.tiles;
	}

	private void initMap() {
		this.gridWidth = 16;
		this.gridHeight = 16; 
		
		this.tiles.grow(16, 16);
		
		for (int x = 0; x < gridWidth; x++) { 
			for (int y = 0; y < gridHeight; y++) {
				this.tiles.set(x, y, new Tile(TextureCache.instanceTiles.getDefault()));
				Logger.messageDebug("Tile: " + x + ":" + y + "  " + this.tiles.get(x, y)); 
			}
		}
	} 

	protected void loadLine(String line) {
		if ((line == null) || line.isEmpty()) {
			return;
		}

		String[] lineBits = line.split(",");

		if (lineBits.length < 3) {
			Logger.messageWarning("Malformed line, not loading: " + Arrays.toString(lineBits));
		}

		int index = 0;
		int x = Integer.parseInt(lineBits[index++]);
		int y = Integer.parseInt(lineBits[index++]);
		String texName = lineBits[index++];
		int rot = Integer.parseInt(lineBits[index++]);
		boolean flipH = Boolean.parseBoolean(lineBits[index++]);
		boolean flipV = Boolean.parseBoolean(lineBits[index++]);
 
		Texture tex = TextureCache.instanceTiles.getTex(texName, rot, flipV, flipH);

		Tile tile = new Tile(tex);
		tile.traversable = Boolean.parseBoolean(lineBits[index++]);
		tile.teleportDestinationGrid = lineBits[index++];
		tile.teleportDestinationX = Integer.parseInt(lineBits[index++]);
		tile.teleportDestinationY = Integer.parseInt(lineBits[index++]);
		tile.teleportDirection = TeleportDirection.fromString(lineBits[index++]);

		try {
			tile.message = lineBits[index++];
		} catch (java.lang.ArrayIndexOutOfBoundsException e) {
		
		}
 
		Logger.messageDebug(tile.toString());

		this.tiles.set(x, y, tile);
	}

	protected void saveImpl(FileWriter fw) {
		for (int x = 0; x < gridWidth; x++) { 
			for (int y = 0; y < gridHeight; y++) {
				this.saveTile(fw, x, y, this.tiles.get(x, y));
			}
		}
	}  

	private void saveTile(FileWriter fw, int x, int y, Tile tile) {
		try {
			fw.append(this.getCsvLine(x, y, tile.getTextureFilenameOnly(), tile.tex.rot, tile.tex.flipH, tile.tex.flipV, tile.traversable, tile.teleportDestinationGrid, tile.teleportDestinationX, tile.teleportDestinationY, tile.teleportDirection, tile.message));
		} catch (IOException e) {
			e.printStackTrace();
		}
	}

	@Override
	public int getGridHeight() {
		return this.gridHeight;
	}

	@Override
	public int getGridWidth() {
		return this.gridWidth; 
	}

	@Override
	public Grid<EntityInstance> getEntityList() {
		return null;   
	}

	@Override
	public int nextEntityId() {
		return 0;
	}

	@Override
	public void grow(int i, int j) {
	}
}
