package greyvarEditor.ui.windows.editors.grid.panels;

import java.awt.Color;
import java.awt.Graphics;
import java.awt.Graphics2D;
import java.awt.Image;
import java.awt.geom.AffineTransform;
import java.awt.image.AffineTransformOp;
import java.awt.image.BufferedImage;

import jwrCommonsJava.Logger;

public class Texture {
	public Image image;
	private final String filename;
	public int rot = 0; 
	public boolean flipV = false;
	public boolean flipH = false;

	public Texture(Image tex, String filename) {
		this.image = tex;
		this.filename = filename;
	}

	@Override
	public Texture clone() {
		BufferedImage bi = new BufferedImage(this.image.getWidth(null), this.image.getHeight(null), BufferedImage.TYPE_INT_ARGB);
		Graphics g = bi.getGraphics(); 
		g.drawImage(this.image, 0, 0, this.image.getWidth(null), this.image.getHeight(null), null);

		return new Texture(bi, this.filename);
	}
	
	public Image getInactive() { 
		BufferedImage bi = new BufferedImage(this.image.getWidth(null), this.image.getHeight(null), BufferedImage.TYPE_INT_ARGB); 
		Graphics g = bi.getGraphics();
		g.setColor(new Color(1f, 1f, 1f, .5f));  
		
		g.drawImage(this.image, 0, 0, this.image.getWidth(null), this.image.getHeight(null), null);
		g.fillRect(0, 0, bi.getWidth(), bi.getHeight());
		
		return bi;
	}

	public void flip(boolean flipV, boolean flipH) {
		this.flipV = flipV;
		this.flipH = flipH;

		BufferedImage bi;

		if (flipV) {
			bi = new BufferedImage(this.image.getWidth(null), this.image.getHeight(null), BufferedImage.TYPE_INT_ARGB);
			bi.getGraphics().drawImage(this.image, 0, 0, null);

			AffineTransform tx = AffineTransform.getScaleInstance(1, -1);
			tx.translate(0, -this.image.getHeight(null));

			this.image = new AffineTransformOp(tx, AffineTransformOp.TYPE_NEAREST_NEIGHBOR).filter(bi, null);
		}

		if (flipH) {
			bi = new BufferedImage(this.image.getWidth(null), this.image.getHeight(null), BufferedImage.TYPE_INT_ARGB);
			bi.getGraphics().drawImage(this.image, 0, 0, null);

			AffineTransform tx = AffineTransform.getScaleInstance(-1, 1);
			tx.translate(-this.image.getWidth(null), 0);

			this.image = new AffineTransformOp(tx, AffineTransformOp.TYPE_NEAREST_NEIGHBOR).filter(bi, null);
		}
	}

	public String getFilename() { 
		return this.filename;
	}

	public String getId() {
		return this.filename + ":" + this.rot + ":" + ((this.flipV) ? 'V' : 'v') + ":" + ((this.flipH) ? 'H' : 'h');
	}

	public void rotate(int rot) {
		this.rot = rot;
		this.rot = (this.rot >= 360) ? this.rot - 360 : this.rot;

		BufferedImage bi = new BufferedImage(this.image.getWidth(null), this.image.getHeight(null), BufferedImage.TYPE_INT_ARGB);

		Graphics2D g2d = (Graphics2D) bi.getGraphics();
		g2d.clearRect(1, 1, bi.getWidth(), bi.getHeight());
		g2d.setColor(Color.PINK);
		g2d.drawRect(0, 0, bi.getWidth(), bi.getHeight());
		g2d.drawImage(this.image, 0, 0, null);

		AffineTransform tx = new AffineTransform();
		tx.rotate(Math.toRadians(rot), this.image.getWidth(null) / 2, this.image.getHeight(null) / 2);

		this.image = new AffineTransformOp(tx, AffineTransformOp.TYPE_BILINEAR).filter(bi, null);
	}

	@Override
	public String toString() {
		return "Texture {" + this.getId() + "}";
	}
}
