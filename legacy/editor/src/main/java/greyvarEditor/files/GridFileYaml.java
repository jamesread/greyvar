package greyvarEditor.files;

import java.beans.Transient;
import java.io.File;
import java.io.FileReader;
import java.io.FileWriter;
import java.io.IOException;
import java.io.StringWriter;

import com.fasterxml.jackson.annotation.JsonInclude;
import com.fasterxml.jackson.core.JsonGenerationException;
import com.fasterxml.jackson.core.JsonGenerator;
import com.fasterxml.jackson.core.JsonLocation;
import com.fasterxml.jackson.core.JsonToken;
import com.fasterxml.jackson.databind.JsonMappingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.dataformat.yaml.UTF8Reader;
import com.fasterxml.jackson.dataformat.yaml.YAMLFactory;
import com.fasterxml.jackson.dataformat.yaml.YAMLParser;

import greyvarEditor.TextureCache;
import greyvarEditor.Tile;
import greyvarEditor.dataModel.EntityInstance;
import jwrCommonsJava.Logger;

public class GridFileYaml implements GridFile {
	private Grid<Tile> tiles = new Grid<>();
	private Grid<EntityInstance> entities = new Grid<>(); 
	private int gridWidth;  
	private int gridHeight; 
	private int lastEntityId = 0;  
	
	private transient File f;
	
	
	@Override
	@Transient  
	public File getFile() { 
		return f;
	}

	public GridFileYaml(File f) {
		this.f = f; 
		 
		this.load();
	}
	
	@Override
	public Grid<Tile> getTileList() {
		return tiles;  
	} 

	@Override 
	public String getFilename() {
		return f.getName(); 
	}

	@Override
	public void load() {
		this.gridWidth = 0;
		this.gridHeight = 0; 
	
		try {
			YAMLFactory f = new YAMLFactory();
			FileReader fr = new FileReader(this.f); 
			YAMLParser p =  f.createParser(fr);
			
			p.nextToken();
			  
			while (p.hasCurrentToken()) { 
				JsonToken t = p.nextToken(); 
				
				if (t == null) {
					break; 
				}
				
				switch (t) {
				case FIELD_NAME:
					switch (p.currentName()) {
						case "tiles":
							loadTileList(p);
							break; 
						case "entities":
							loadEntityList(p);
							break;
						case "width":
						case "colcount":
							this.grow(p.nextIntValue(16), 0); // growing from 0, 0 initial size
							break;  
						case "height":
						case "rowcount":
							this.grow(0, p.nextIntValue(16)); // growing from 0, 0 initial size
							break; 
						case "lastEntityId":
							this.lastEntityId = p.nextIntValue(0);
							break;  
						default: 
							Logger.messageDebug("Unhandled field: " + p.currentName());
					} 
					 
					break; 
				case VALUE_STRING:
					break; 
				case END_OBJECT: 
					break;
				default:
					this.logUnhandledToken(t, p);
				}
			}
		} catch (Exception e) {
			e.printStackTrace();
		}

		if (this.gridWidth == 0) {
			this.grow(4, 4);
		}
	}

	private void logUnhandledToken(JsonToken t, YAMLParser p) {
		JsonLocation loc = p.getCurrentLocation(); 
		 
		System.out.println("Unhandled token: " + t + "(" + p.getCurrentValue() + "). line: " + loc.getLineNr() + " pos " + loc.getColumnNr()); 
	}
	
	private void loadEntityList(YAMLParser p)  {
		boolean endOfList = false;
		
		int row = -1;
		int col = -1;
		int id = -1; 
		String definition = "";
		
		try {
			while (p.hasCurrentToken()) {
				JsonToken t = p.nextToken();
				
				if (t == null) {
					break; 
				}
				
				switch(t) {
				case START_ARRAY:
					break;
				case START_OBJECT:
					row = -1;
					col = -1;
					id = -1;
					definition = "";
					
					break;
				case END_ARRAY: 
					endOfList = true;
				case END_OBJECT:   
					if (row == -1 || col == -1) {
						continue;
					} 
					
					EntityInstance e = new EntityInstance(definition, id);    
					
					this.entities.set(row, col, e);
					break;
				case FIELD_NAME:
					switch(p.currentName()) {
					case "row":
					case "x": 
						row = p.nextIntValue(-1);
						break;
					case "col":
					case "y": 
						col = p.nextIntValue(-1);
						break;
					case "id": 
						id = p.nextIntValue(-1);
						break;
					case "definition": 
						definition = p.nextTextValue();
						break; 
					}
				}
			}
		} catch (Exception e) {
			e.printStackTrace(); 
		}
	}

	private void loadTileList(YAMLParser p) {
		boolean endOfList = false;
		
		int x = -1;
		int y = -1;
		int rot = 0;
		boolean traversable = false;
		boolean flipH = false;
		boolean flipV = false;
		String texture = ""; 
		
		try { 
			while (p.hasCurrentToken()) {
				JsonToken t = p.nextToken();
				
				switch (t) { 
				case START_ARRAY:
					break;
				
				case START_OBJECT: 
					x = -1; 
					y = -1; 
					rot = 0;
					flipH = false; 
					flipV = false;  
					traversable = false;
					texture = "";
					
					break; 
				case END_ARRAY:  
					endOfList = true;
					// fall through to end the last object
				case END_OBJECT:  
					Tile tile = new Tile(TextureCache.instanceTiles.getTex(texture, rot, flipV, flipH));
					tile.traversable = traversable;
					
					this.tiles.set(x, y, tile);
					break;
				case FIELD_NAME:
					switch (p.currentName()) {
					case "traversable": 
						traversable = p.nextBooleanValue();
						break; 
					case "x":
						x = p.nextIntValue(-1);
						break;
					case "y":
						y = p.nextIntValue(-1);
						break;
					case "flipH":
						flipH = p.nextBooleanValue();
						break;
					case "flipV":
						flipV = p.nextBooleanValue();
						break;
					case "texture":
						texture = p.nextTextValue();
						break;
					case "rot":
						rot = p.nextIntValue(-1);
						break; 
					default:
						Logger.messageDebug("Unhandled tile field: " + p.currentName()); 
					} 
					
					break;
				default:
					this.logUnhandledToken(t, p); 
				}
				
				if (endOfList) {
					return;
				}
			}
		} catch (Exception  e) {
			e.printStackTrace();;
		}
	}

	@Override
	public int getGridHeight() {
		return gridHeight;
	}
 
	@Override
	public int getGridWidth() {
		return gridWidth;
	}

	@Override
	public void save() {
		try {
			YAMLFactory f = new YAMLFactory();
			FileWriter fw = new FileWriter(this.f); 
			JsonGenerator gen = f.createGenerator(fw);
			 
			gen.writeStartObject();
			gen.writeNumberField("width", this.gridWidth);
			gen.writeNumberField("height", this.gridHeight);
			gen.writeNumberField("lastEntityId", this.lastEntityId);
			
			gen.writeFieldName("tiles");
			gen.writeStartArray();
			for (int x = 0; x < gridWidth; x++) {
				for (int y = 0; y < gridHeight; y++) {
					Tile t = this.tiles.get(x, y);  
					
					gen.writeStartObject();					
					gen.writeStringField("texture", t.getTextureFilenameOnly());
					gen.writeNumberField("x", x);
					gen.writeNumberField("y", y);
					gen.writeNumberField("rot", t.tex.rot);
					gen.writeBooleanField("flipH", t.tex.flipH);
					gen.writeBooleanField("flipV", t.tex.flipV);
					gen.writeBooleanField("traversable", t.traversable);
					 
					if (t.teleportDestinationGrid != null && !t.teleportDestinationGrid.isEmpty()) {
						gen.writeStringField("teleportDst", t.teleportDestinationGrid);
						gen.writeStringField("teleportDir", t.teleportDirection.toString()); 
						gen.writeNumberField("teleportX", t.teleportDestinationX);
						gen.writeNumberField("teleportY", t.teleportDestinationY); 
					} 					
					
					gen.writeEndObject(); 
				} 
			} 
			gen.writeEndArray();
			
			gen.writeFieldName("entities");
			gen.writeStartArray();
			for (int x = 0; x < gridWidth; x++) {
				for (int y = 0; y < gridHeight; y++) {
					EntityInstance e = this.entities.get(x, y);
 
					if (e == null || e.editorTexture == null) {
						continue;
					}
					
					gen.writeStartObject();
					gen.writeNumberField("x", x);
					gen.writeNumberField("y", y);   
					gen.writeNumberField("id", e.id);  
					gen.writeStringField("definition", this.entities.get(x, y).definition);
					
					gen.writeEndObject();
				}
			}
			gen.writeEndArray();
 		
			fw.close();
		} catch (Exception e) {
			e.printStackTrace();
		} 
	}

	public void setTileList(Grid<Tile> tileList) {
		this.tiles = tileList;		
	}

	@Override 
	public Grid<EntityInstance> getEntityList() {
		return this.entities; 
	}

	@Override 
	public int nextEntityId() {
		this.lastEntityId++;
		
		return this.lastEntityId; 
	} 

	@Override
	public void grow(int w, int h) {
		for (Grid.Pos p : this.tiles.grow(w, h)) {
			this.tiles.set(p, new Tile(TextureCache.instanceTiles.getDefault()));		
		}
			
		this.entities.grow(w, h);

		this.gridWidth += w;
		this.gridHeight += h;
	}

}
