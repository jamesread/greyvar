package greyvarEditor.files;

import java.io.File;

import greyvarEditor.Tile;
import greyvarEditor.dataModel.EntityInstance;

public interface GridFile { 
	public Grid<Tile> getTileList();

	public String getFilename(); 
	
	public void load();

	public int getGridHeight();
 
	public int getGridWidth();

	public void save();

	public File getFile(); 
 
	public Grid<EntityInstance> getEntityList();   
	
	public int nextEntityId();

	public void grow(int i, int j);  
}
 