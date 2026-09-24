package greyvarEditor;

import java.awt.Color;
import java.awt.Image;
import java.awt.Graphics2D;
import java.awt.image.BufferedImage;
import java.io.File;
import java.util.HashMap;

import javax.imageio.ImageIO;

import greyvarEditor.ui.windows.editors.grid.panels.Texture;
import jwrCommonsJava.Logger;

public class TextureCache {
	private class MutatedTexture {
		public Texture texOriginal;
		public int rot;
		public boolean flipV;
		public boolean flipH;

		public MutatedTexture(Texture t, int rot, boolean flipV, boolean flipH) {
			this.texOriginal = t;
			this.rot = rot;
			this.flipV = flipV;
			this.flipH = flipH;
		}

		@Override
		public boolean equals(Object comp) {
			if (comp instanceof MutatedTexture) {
				MutatedTexture comp2 = (MutatedTexture) comp;

				if (!(this.texOriginal.equals(comp2.texOriginal))) {
					return false;
				}

				if (this.rot != comp2.rot) {
					return false;
				}

				if (this.flipV != comp2.flipV) {
					return false;
				}

				if (this.flipH != comp2.flipH) {
					return false;
				}

				return true;
			} else {
				return super.equals(comp);
			}
		}
	}

	private final HashMap<String, Texture> textureCacheOriginals = new HashMap<String, Texture>();
	private final HashMap<MutatedTexture, Texture> textureCacheMutated = new HashMap<MutatedTexture, Texture>();

	public static TextureCache instanceTiles = new TextureCache("tiles");
	public static TextureCache instanceEntities = new TextureCache("entities");
	public static TextureCache instanceFluids = new TextureCache("fluids");

	private Texture defaultTexture;
	File baseDir;

	private TextureCache(String subdir) {	
		this.baseDir = new File(GameResources.dirTextures, subdir); 

		Logger.messageDebug("texture cache for " + subdir + " will use base dir of; " + this.baseDir.getAbsolutePath());

		this.defaultTexture = this.loadTexConstruct();
	}

	public HashMap<String, Texture> getAll() {
		return this.textureCacheOriginals;
	}

	public Texture getDefault() {
		return this.defaultTexture;
	}

	public Texture getTex(String name) {
		return this.getTex(name, 0);     
	}
	
	private Texture getOriginalTexture(String name) {
		if (this.textureCacheOriginals.containsKey(name)) {
			return this.textureCacheOriginals.get(name);
		} else {
			return this.loadTex(name);
		}
	}
	
	public Texture getTex(String name, int rot) {  
		return this.getTex(name, rot, false, false); 
	}
 
	public Texture getTex(String name, int rot, boolean flipV, boolean flipH) { 
		Texture t = this.getOriginalTexture(name);

		if ((rot == 0) && !flipH && !flipV) {
			return t;
		} else {
			MutatedTexture mutatedAttributes = new MutatedTexture(t, rot, flipV, flipH);

			if (this.textureCacheMutated.containsKey(mutatedAttributes)) {
				return this.textureCacheMutated.get(mutatedAttributes);
			} else {
				Texture mutatedTexture = t.clone();
				mutatedTexture.flip(mutatedAttributes.flipV, mutatedAttributes.flipH);
				mutatedTexture.rotate(rot);

				this.textureCacheMutated.put(mutatedAttributes, mutatedTexture);

				return mutatedTexture;
			}
		}
	} 
	
	private Texture loadTex(File f) {
		return this.loadTex(f.getName()); 
	}

	private Texture loadTex(String texName) {
		if (texName.isEmpty()) {
			return this.defaultTexture;
		}

		if (this.textureCacheOriginals.containsKey(texName)) {
			Logger.messageDebug("Replacing texture filename; " + texName);
			this.textureCacheOriginals.remove(texName);
		}

		try {
			File f = new File(baseDir + "/" + texName);
			Image i = ImageIO.read(f);
			Texture tex = new Texture(i, f.getName());
   
			Logger.messageDebug("Loaded tex (" + this.baseDir.getName() + " cache): " + texName);
			this.textureCacheOriginals.put(texName, tex);

			return tex;
		} catch (Exception e) {
			Logger.messageWarning("Failed to load: " + texName + " from " + this.baseDir + " because: " + e.getMessage());
			this.textureCacheOriginals.put(texName, this.defaultTexture);
		}
		
		return this.defaultTexture; 
	}

	public Texture loadTexConstruct() {
		final int w = 32;
		final int h = 32;

		BufferedImage bi = new BufferedImage(w, h, BufferedImage.TYPE_INT_RGB);
		Graphics2D g2d = bi.createGraphics();
		g2d.setColor(Color.orange);
		g2d.fillRect(0, 0, w, h);
		g2d.dispose();

		Texture tex = new Texture(bi, "<construct>");

		return tex;
	}
  
	public void loadTextures() {
		Logger.messageDebug("Loading textures from directory: " + baseDir.getAbsolutePath());
 
		for (File f : baseDir.listFiles()) {
			if (f.getName().endsWith(".png")) {
				this.loadTex(f.getName());  
			}
		}
	}
}
