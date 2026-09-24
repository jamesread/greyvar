package greyvarEditorTests.gridLoading;

import java.util.Vector;

import greyvarEditor.files.Grid;
import junit.framework.TestCase;

public class TestGrid extends TestCase {
	public void testGrid() {
		Vector<Grid.Pos> newCells;
		
		Grid<String> g = new Grid<>();  
		
		assertEquals(g.getWidth(), 0);
		assertEquals(g.getHeight(), 0);
		
		newCells = g.grow(2, 2); 
		
		assertEquals(newCells.size(), 4); 
		assertEquals(g.getWidth(), 2);
		assertEquals(g.getHeight(), 2);
		
		g.set(0, 0, "cake");
		g.set(1, 1, "two"); 
		
		assertEquals(g.get(0, 0), "cake"); 
		assertEquals(g.get(1, 1), "two");
		
		newCells = g.grow(2, 2); 
		  
		assertEquals(newCells.size(), 12);
		assertEquals(g.getWidth(), 4);
		assertEquals(g.getHeight(), 4); 
		 
		assertEquals(g.get(0, 0), "cake"); 
		assertEquals(g.get(1, 1), "two");
		assertEquals(g.get(0, 0), "cake"); 
		assertEquals(g.get(2, 2), null); 
	}
}
