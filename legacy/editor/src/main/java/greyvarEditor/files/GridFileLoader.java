package greyvarEditor.files;

import java.io.BufferedReader;
import java.io.File;
import java.io.FileNotFoundException;
import java.io.FileReader;

import greyvarEditor.Tile;

public abstract class GridFileLoader {
	public static GridFile load(String s) {
		return load(new File(s)); 
	} 
	
	public static GridFile load(File f) {
		GridFile gf;
		
		BufferedReader br;
		try {
			br = new BufferedReader(new FileReader(f));
			
			String firstLine = br.readLine();
			
			 
			if (firstLine != null && firstLine.startsWith("---")) { 
				gf = new GridFileYaml(f);			
			} else {
				gf = new GridFileCsv(f);
			}
			
			br.close();
		} catch (Exception e) {
			e.printStackTrace(); 
			return null;
		}
		
		return gf; 
	}
	
	public static GridFileYaml toYaml(GridFile gf) {
		File f2 = new File(gf.getFile().getAbsolutePath().replace(".csv", ".grid"));
		
		GridFileYaml ret = new GridFileYaml(f2);   
		
		ret.setTileList(gf.getTileList());
		
		return ret;
	}
}
