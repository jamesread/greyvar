package greyvarEditor;

import java.io.File;

import greyvarEditor.ui.windows.editors.grid.panels.Texture;
import jwrCommonsJava.Logger;

public class Tile {
	public enum TeleportDirection {
		NORTH, EAST, SOUTH, WEST;

		public static TeleportDirection fromString(String string) {
			if (string.equals("NORTH")) {
				return NORTH;
			} else if (string.equals("EAST")) {
				return EAST;
			} else if (string.equals("SOUTH")) {
				return SOUTH;
			} else if (string.equals("WEST")) {
				return WEST;
			} else {
				throw new IllegalArgumentException("Direction cannot be parsed.");
			}
		}

		public int getDegrees() {
			switch (this) {
			case NORTH:
				return 0;
			case EAST:
				return 90;
			case SOUTH:
				return 180;
			case WEST:
				return 270; 
			default:
				return 0;
			}
		}
	};

	public transient Texture tex;
	public boolean traversable = true;
	public int teleportDestinationX = 0;
	public int teleportDestinationY = 0;
	public String teleportDestinationGrid = "";
	public TeleportDirection teleportDirection = TeleportDirection.SOUTH;
	public boolean hover;
	public String message;

	public Tile(Texture tex) {
		if (tex == null) {
			throw new NullPointerException("Texture for the tile cannot be null");
		}

		this.tex = tex;
	}

	public String getTextureFilenameOnly() {
		return this.tex.getFilename();
	}

	@Override  
	public String toString() {
		return "Tile {tex: " + this.tex.toString() + "}";
	}
}
