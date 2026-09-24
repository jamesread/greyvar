package greyvarEditor.ui.components;

import greyvarEditor.TextureCache;
import greyvarEditor.Tile;
import greyvarEditor.dataModel.EntityInstance;
import greyvarEditor.files.GridFile;
import greyvarEditor.ui.windows.WindowEditorGrid;
import greyvarEditor.ui.windows.editors.grid.panels.Texture;
import greyvarEditor.utils.EditLayerMode;

import java.awt.Color;
import java.awt.Dimension;
import java.awt.Graphics;
import java.awt.Graphics2D;
import java.awt.Image;
import java.awt.Point;
import java.awt.RenderingHints; 
import java.awt.event.MouseEvent;
import java.awt.event.MouseListener;
import java.awt.event.MouseMotionListener;
import java.awt.image.BufferedImage;
import java.util.Iterator;
import java.util.Vector;

import javax.swing.JComponent;

public class ComponentGridEditor extends JComponent implements MouseListener, MouseMotionListener {
	private EditLayerMode currentEditMode = EditLayerMode.TILES;  

	private int currx;
	private int curry;

	private GridFile gf;
	private Tile lastHoveredTile;
	private int selectionRectSize = 1;
	public boolean highlightTeleports = false;
	public boolean highlightTraversable = false;
	   
	private final WindowEditorGrid windowEditorGrid; 

	public ComponentGridEditor(WindowEditorGrid windowEditorGrid) {
		this.windowEditorGrid = windowEditorGrid; 
		
		this.addMouseListener(this);
		this.addMouseMotionListener(this);
		this.setPreferredSize(new Dimension(50, 50));
	}
	
	@Override
	public Dimension getPreferredSize() {    
		return new Dimension(gf.getGridWidth() * 64, gf.getGridHeight() * 64);
	}
	
	@Override
	public Dimension getMaximumSize() {
		return getPreferredSize(); 
	}
	
	@Override
	public Dimension getMinimumSize() {
		return getPreferredSize(); 
	}	

	public void blanketTextureEverything(Texture t) {
		for (int i = 0; i < this.gf.getGridWidth(); i++) {
			for (int j = 0; j < this.gf.getGridHeight(); j++) {
				this.gf.getTileList().get(i, j).tex = t;
			} 
		}
 
		this.repaint();
	}

	public GridFile getGridFile() {
		return this.gf;
	}

	public Image getScreenshot() {
		BufferedImage bi = new BufferedImage(this.getWidth(), this.getHeight(), BufferedImage.TYPE_4BYTE_ABGR);

		Graphics2D g = (Graphics2D) bi.getGraphics();

		this.paint(g);

		return this.scaleImage(bi, 64, 64, Color.WHITE);
	}

	public Tile getTileAt(int x, int y) { 
		return this.gf.getTileList().get(x, y);
	}

	public Point getCellAt(Point p) {
		int w = this.getWidth() / gf.getGridWidth();
		int h = this.getHeight() / gf.getGridHeight();

		int x = (int) p.getX() / w;
		int y = (int) p.getY() / h;

		return new Point(x, y);
	}
 
	public Iterable<Tile> getTiles() {
		return gf.getTileList().getIterator();
	}

	@Override
	public void mouseClicked(MouseEvent e) {

	}

	@Override
	public void mouseDragged(MouseEvent e) {
		// TODO Auto-generated method stub

	}

	@Override
	public void mouseEntered(MouseEvent e) {
	}

	@Override
	public void mouseExited(MouseEvent e) {
	}

	@Override
	public void mouseMoved(MouseEvent e) {
		Point p = this.getCellAt(e.getPoint());

		if ((p.x < gf.getGridWidth()) && (p.y < gf.getGridHeight())) {
			Tile t = this.gf.getTileList().get(p.x, p.y); 

			if (this.lastHoveredTile != null) {
				this.lastHoveredTile.hover = false;
			}

			if (this.lastHoveredTile != t) {
				t.hover = true;
				this.lastHoveredTile = t;
				this.repaint();
			}
		}
	}

	@Override 
	public void mousePressed(MouseEvent e) {
		Point p = this.getCellAt(e.getPoint());

		this.currx = p.x;
		this.curry = p.y;

		if (e.getButton() == MouseEvent.BUTTON1) {
			if (e.getClickCount() > 1) { 
				this.windowEditorGrid.onCellSelected(this.currx, this.curry, this.gf.getTileList().get(p.x, p.y), this.gf.getEntityList().get(p.x, p.y));
			} else {
				this.windowEditorGrid.onCellFocus(this.currx, this.curry, this.gf.getTileList().get(p.x, p.y), this.gf.getEntityList().get(p.x, p.y));
			}
		} else if (e.getButton() == MouseEvent.BUTTON3) { 
			this.onCellApplyPaint(this.currx, this.curry);
		} 

		this.repaint(); 
	} 
	
	private void onCellApplyPaint(int x, int y) {
		switch (this.currentEditMode) {
		case ENTITIES:  
			EntityInstance enti = windowEditorGrid.panEntity.getNewSelected(this.getGridFile().nextEntityId());
			 
			this.gf.getEntityList().set(x, y, enti);
			break;
		case TILES: 
			this.gf.getTileList().get(x, y).tex = windowEditorGrid.panAppearance.getCurrentTexture();
		} 
	}

	@Override
	public void mouseReleased(MouseEvent e) {
	}

	@Override
	public void paint(Graphics g) {
		int w = this.getWidth() / gf.getGridWidth();
		int h = this.getHeight() / gf.getGridHeight(); 

		for (int x = 0; x < gf.getGridWidth(); x++) {
			for (int y = 0; y < gf.getGridHeight(); y++) {
				this.paintTile(g, x, y, this.gf.getTileList().get(x, y), w, h);
			}
		} 
		
		if (this.getEditLayerMode() == EditLayerMode.ENTITIES) {
			g.setColor(new Color(255, 255, 255, 100));   
			g.fillRect(0, 0, this.getWidth(), this.getHeight());  
		}  
		
		for (int x = 0; x < gf.getGridWidth(); x++) {
			for (int y = 0; y < gf.getGridHeight(); y++) { 
				this.paintEntity(g, x, y, this.gf.getEntityList().get(x, y), w, h);
			}
		}  
	} 
	
	public void paintEntity(Graphics g, int x, int y, EntityInstance e, int w, int h) {
		if (e == null || e.editorTexture == null) {
			return;
		} else {
			g.drawImage(e.editorTexture.image, x * w, y * h, w, h, null); 
		}
	} 

	public void paintTile(Graphics g, int x, int y, Tile t, int w, int h) {
		if (t == null) {
			g.setColor(Color.PINK);
			g.fillRect(x * w, y * h, w, h);
		} else {
			g.drawImage(t.tex.image, x * w, y * h, w, h, null);
		}

		if ((x == this.currx) && (y == this.curry)) {
			g.setColor(Color.RED);
			g.drawRect(x * w, y * h, w - 1, h - 1);
		} else if (t.hover) {
			g.setColor(Color.BLUE);
			g.drawRect(x * w, y * h, (w * this.selectionRectSize) - 1, (h * this.selectionRectSize) - 1);
		}

		if (this.highlightTeleports && !t.teleportDestinationGrid.isEmpty()) {
			g.setColor(Color.ORANGE);
			g.fillOval((x * w) + (w / 4), (y * h) + (h / 4), w / 2, h / 2);
			Texture i = TextureCache.instanceEntities.getTex("arrow.png", t.teleportDirection.getDegrees()); 
			g.drawImage(i.image, (x * w) + (w / 4), (y * h) + (h / 4), w / 2, h / 2, null);
		}

		if (this.highlightTraversable) {
			if (t.traversable) {
				g.setColor(Color.GREEN);
			} else {
				g.setColor(Color.RED);
			}

			g.fillRect((x * w) + (w / 4), (y * h) + (h / 4), w - ((w / 4) * 2), h - ((h / 4) * 2));
		}
	}

	public BufferedImage scaleImage(BufferedImage img, int width, int height, Color background) {
		int imgWidth = img.getWidth();
		int imgHeight = img.getHeight();
		if ((imgWidth * height) < (imgHeight * width)) {
			width = (imgWidth * height) / imgHeight;
		} else {
			height = (imgHeight * width) / imgWidth;
		}
		BufferedImage newImage = new BufferedImage(width, height, BufferedImage.TYPE_INT_RGB);
		Graphics2D g = newImage.createGraphics();
		try {
			g.setRenderingHint(RenderingHints.KEY_INTERPOLATION, RenderingHints.VALUE_INTERPOLATION_BICUBIC);
			g.setBackground(background);
			g.clearRect(0, 0, width, height);
			g.drawImage(img, 0, 0, width, height, null);
		} finally {
			g.dispose();
		}
		return newImage;
	}

	public void setGridFile(GridFile gf) {
		this.gf = gf;
		this.repaint(); 
	}

	public void setSelectionRectSize(int size) {
		this.selectionRectSize = size;
	}

	public void setTileAt(int x, int y, Tile tile) {
		this.gf.getTileList().set(x, y, tile);
	} 

	public void setEditMode(EditLayerMode mode) {
		this.currentEditMode = mode;
	}
  
	public EditLayerMode getEditLayerMode() {
		return this.currentEditMode;
	}
}
