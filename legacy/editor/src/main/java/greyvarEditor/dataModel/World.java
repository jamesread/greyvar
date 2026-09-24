package greyvarEditor.dataModel;

import java.io.File;
import java.util.Vector;

import greyvarEditor.triggers.Trigger;

public class World {
	public String author = "unknown";
	public String title = "unknown"; 
	public String spawnGrid = "unknown";  
	
	public transient Vector<File> gridFiles = new Vector<File>();
	public transient File file;
	
	public Vector<Trigger> triggers = new Vector<>();
}
