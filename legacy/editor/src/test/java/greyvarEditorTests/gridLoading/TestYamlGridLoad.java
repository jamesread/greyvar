package greyvarEditorTests.gridLoading;

import java.io.File;

import greyvarEditor.Tile;
import greyvarEditor.dataModel.EntityInstance;
import greyvarEditor.files.Grid;
import greyvarEditor.files.GridFileYaml;
import junit.framework.TestCase;
import jwrCommonsJava.Configuration;

import static org.junit.Assert.*;
 
public class TestYamlGridLoad extends TestCase {
	public void testLoad() { 
		Configuration.set("sysAppName", "greyvarEditor");
		Configuration.parseUserConfiguration();
		
		GridFileYaml gridFile = new GridFileYaml(new File("src/test/testFiles/0.0.grid"));
		
		assertEquals(16, gridFile.getGridWidth());
		assertEquals(16, gridFile.getGridHeight()); 
		
		Grid<Tile> tileList = gridFile.getTileList();
		
		assertNotNull(tileList);    
		
		Grid<EntityInstance> entityList = gridFile.getEntityList();
		
		System.out.println(entityList); 
		   
		//assertEquals("hill.png", tileList.get(0, 0).getTextureFilenameOnly()); // If textures are unavailable, the tex will be "<construct>", which is OK
		//assertEquals("chest", entityList.get(13, 4).definition);   
	}
} 
