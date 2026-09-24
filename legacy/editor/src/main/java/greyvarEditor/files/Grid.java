package greyvarEditor.files;

import java.util.Collection;
import java.util.HashMap;
import java.util.Iterator;
import java.util.List;
import java.util.Vector;

import greyvarEditor.Tile;

public class Grid<T> {
	private HashMap<Integer, HashMap<Integer, T>> grid = new HashMap<Integer, HashMap<Integer, T> >();
	
	public static class Pos {
		int x = 0;
		int y = 0;  
		
		public Pos(int x, int y) {
			this.x = x;
			this.y = y; 
		}
	}
	
	int width = 0;
	int height = 0;
	
	public Grid() { 
	}
	
	public Grid(int w, int h) {
		 this.grow(w, h);
	}
	
	public void set(Pos p, T object) {
		this.set(p.x, p.y, object); 
	}
  
	public void set(int x, int y, T object) {
		this.boundsCheck(x, y);
		
		this.grid.get(x).put(y, object);
	}  

	public T get(int x, int y) { 
		this.boundsCheck(x, y);
		
		return this.grid.get(x).get(y);
	}
	 
	private void boundsCheck(int x, int y) {
		if (x >= width) {
			throw new ArrayIndexOutOfBoundsException(x + " is larger than width " + width);
		}
		 
		if (y >= height) {
			throw new ArrayIndexOutOfBoundsException(y + " is larger than height " + height);
		} 
	}

	public Vector<Pos> grow(int growX, int growY) {
		Vector<Pos> newCells = new Vector<>();
		
		for (int x = 0; x < width; x++) { 
			for (int y = height; y < height + growY; y++) {
				this.grid.get(x).put(y, null);
				newCells.add(new Pos(x, y));
			} 
		}
		
		for (int x = width; x < width + growX; x++) {
			this.grid.put(x, new HashMap<Integer, T>());
			
			for (int y = 0; y < height + growY; y++) {
				this.grid.get(x).put(y, null);
				newCells.add(new Pos(x, y));
			} 
		}   
		
		this.width += growX;   
		this.height += growY;	
		
		return newCells;
	}
  
	public Iterable<T> getIterator() {
		List<T> all = new Vector<>();
		 
		for (HashMap<Integer, T> x : grid.values()) {
			all.addAll(x.values());  
		}
		
		return all; 
	}

	public int getHeight() {
		return this.height;
	}
	
	public int getWidth() {
		return this.width; 
	}
	
	@Override
	public String toString() {
		String s = "";
		 
		s += "grid {" + this.width + ":" + this.height + "}\n"; 
		
		for (int k : this.grid.keySet()) {
			for (T t : this.grid.get(k).values()) {
				s += t;
				s += ", ";   
			}
			
			s += "\n";
		}
		
		return s;
	}
}
