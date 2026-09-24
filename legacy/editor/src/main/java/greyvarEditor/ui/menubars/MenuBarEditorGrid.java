package greyvarEditor.ui.menubars;

import greyvarEditor.TextureCache;
import greyvarEditor.files.GridFile;
import greyvarEditor.files.GridFileLoader;
import greyvarEditor.ui.windows.PopableWindow;
import greyvarEditor.ui.windows.WindowEditorGrid;
import greyvarEditor.ui.windows.WindowFindReplace;
import greyvarEditor.ui.windows.WindowResizeGrid;
import greyvarEditor.ui.windows.editors.grid.panels.Texture;
import greyvarEditor.utils.EditLayerMode;
import greyvarEditor.utils.TextureProertiesReader;

import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.util.Map;

import javax.swing.JCheckBoxMenuItem;
import javax.swing.JMenu;
import javax.swing.JMenuBar;
import javax.swing.JMenuItem;

import jwrCommonsJava.Logger;

public class MenuBarEditorGrid extends JMenuBar implements ActionListener, PopableWindow.Listener {
	private final JMenuItem mniSave = new JMenuItem("Save");
	private final JMenuItem mniResize = new JMenuItem("Resize"); 
	private final JMenuItem mniDebug = new JMenuItem("Debug");
	private final JMenuItem mniReload = new JMenuItem("Reload grid from disk");
	private final JMenuItem mniPopper = new JMenuItem("Pop");
	private final JMenu mnuRectSize = new JMenu("Selection rect size");
	private final JMenuItem mniRectSize1x1 = new JMenuItem("1x1");
	private final JMenuItem mniRectSize3x3 = new JMenuItem("3x3");
	private final WindowEditorGrid windowEditorGrid;
	private final JMenuItem mniBlanketTextureEverything = new JMenuItem("Blanket texture everything"); 
	
	private final JMenu mnuLayers = new JMenu("Layers");
	private final JMenuItem mniEditLayerModeTiles = new JMenuItem("1. Tiles");
	private final JMenuItem mniEditLayerModeEntities = new JMenuItem("2. Entities");
	private final JMenuItem mniEditLayerModeFluids = new JMenuItem("3. Fluids");  
	
	private final JMenu mnuHighlight = new JMenu("Highlight"); 
	private final JCheckBoxMenuItem mniHighlightTeleports = new JCheckBoxMenuItem("Telports");
	private final JCheckBoxMenuItem mniHighlightTraversable = new JCheckBoxMenuItem("Traversable");
	private final JMenuItem mniLoadTextureAttributes = new JMenuItem("Load texture attributes");
	private final JMenu mnuEditor = new JMenu("Editor");
	private final JMenu mnuTextures = new JMenu("Textures");
	private final JMenuItem mniFindReplace = new JMenuItem("Find/Replace");
	private final JMenuItem mniReloadTextures = new JMenuItem("Reload textures, and reload grid");

	public MenuBarEditorGrid(WindowEditorGrid windowEditorGrid) {
		this.windowEditorGrid = windowEditorGrid;
		windowEditorGrid.addListener(this);
		this.setupComponents();
	} 

	@Override
	public void actionPerformed(ActionEvent e) {
		Object source = e.getSource();

		if (source == this.mniSave) {
			this.onMniSaveClicked();
		} else if (source == this.mniDebug) {
			Map<String, Texture> textures = TextureCache.instanceTiles.getAll();

			for (String texName : textures.keySet()) {
				Logger.messageDebug(texName + ": " + textures.get(texName));
			}
		} else if (source == this.mniReload) {
			this.windowEditorGrid.reload();
		} else if (source == mniResize) { 
			WindowResizeGrid resizer = new WindowResizeGrid(this.windowEditorGrid);
			resizer.setVisible(true); 
		} else if (source == this.mniPopper) {
			this.windowEditorGrid.pop();
		} else if (source == this.mniRectSize1x1) {
			this.windowEditorGrid.gridEditor.setSelectionRectSize(1);
		} else if (source == this.mniRectSize3x3) {
			this.windowEditorGrid.gridEditor.setSelectionRectSize(3);
		} else if (source == this.mniBlanketTextureEverything) {
			this.windowEditorGrid.blanketTextureEverything();
		} else if (source == this.mniEditLayerModeTiles) {
			this.windowEditorGrid.setEditMode(EditLayerMode.TILES);
			this.toggleLayerButtons();
		} else if (source == this.mniEditLayerModeFluids) { 
			this.windowEditorGrid.setEditMode(EditLayerMode.FLUIDS);
			this.toggleLayerButtons();
		} else if (source == this.mniEditLayerModeEntities) {
			this.windowEditorGrid.setEditMode(EditLayerMode.ENTITIES);
			this.toggleLayerButtons();
		} else if (source == this.mniHighlightTeleports) {
			this.windowEditorGrid.gridEditor.highlightTeleports = this.mniHighlightTeleports.getState();
			this.windowEditorGrid.gridEditor.repaint();
		} else if (source == this.mniHighlightTraversable) {
			this.windowEditorGrid.gridEditor.highlightTraversable = this.mniHighlightTraversable.getState();
			this.windowEditorGrid.gridEditor.repaint();
		} else if (source == this.mniLoadTextureAttributes) {
			new TextureProertiesReader();
		} else if (source == this.mniReloadTextures) {		
			TextureCache.instanceTiles.loadTextures();
			TextureCache.instanceEntities.loadTextures();   
			this.windowEditorGrid.reload();
		} else if (source == this.mniFindReplace) {
			new WindowFindReplace(this.windowEditorGrid);
		}
	}
	
	public void toggleLayerButtons() {
		this.mniEditLayerModeTiles.setEnabled(true); 
		this.mniEditLayerModeEntities.setEnabled(true); 
		this.mniEditLayerModeFluids.setEnabled(true); 
		
		switch (this.windowEditorGrid.gridEditor.getEditLayerMode()) {
		case TILES: 
			this.mniEditLayerModeTiles.setEnabled(false);
			break;
		case ENTITIES:
			this.mniEditLayerModeEntities.setEnabled(false); 
			break;
		case FLUIDS:  
			this.mniEditLayerModeFluids.setEnabled(false); 
			break; 
		}
	}

	private void onMniSaveClicked() {
		GridFile gf = this.windowEditorGrid.getGridFile();
		
		if (gf.getClass().getSimpleName().endsWith("Csv")) {
			Logger.messageDebug("Changing Csv to Yaml");
			
			gf = GridFileLoader.toYaml(gf);
		}
		 
		gf.save(); 
	}

	private void setupComponents() {
		this.mniSave.addActionListener(this);
		this.mnuEditor.add(this.mniSave);

		this.mniDebug.addActionListener(this);
		this.mnuEditor.add(this.mniDebug);

		this.mniReload.addActionListener(this);
		this.mnuEditor.add(this.mniReload);
		
		this.mniResize.addActionListener(this); 
		this.mnuEditor.add(this.mniResize);

		this.mnuEditor.addSeparator();

		this.mniPopper.addActionListener(this);
		this.mnuEditor.add(this.mniPopper);

		this.mniRectSize1x1.addActionListener(this);
		this.mnuRectSize.add(this.mniRectSize1x1);

		this.mniRectSize3x3.addActionListener(this);
		this.mnuRectSize.add(this.mniRectSize3x3);

		this.mnuEditor.add(this.mnuRectSize);

		this.add(this.mnuEditor);

		this.setupHighlightMenu();
		this.setupLayerMenu();
		this.setupTexturesMenu();
	}

	private void setupHighlightMenu() {
		this.mniHighlightTeleports.addActionListener(this);
		this.mnuHighlight.add(this.mniHighlightTeleports);

		this.mniHighlightTraversable.addActionListener(this);
		this.mnuHighlight.add(this.mniHighlightTraversable);

		this.add(this.mnuHighlight);
	} 
	
	private void setupLayerMenu() {
		this.mniEditLayerModeTiles.addActionListener(this); 
		this.mnuLayers.add(mniEditLayerModeTiles); 
		
		this.mniEditLayerModeEntities.addActionListener(this);
		this.mnuLayers.add(mniEditLayerModeEntities);
		
		this.mniEditLayerModeFluids.addActionListener(this);
		this.mnuLayers.add(mniEditLayerModeFluids); 
		this.add(this.mnuLayers);
	}

	private void setupTexturesMenu() {
		this.mniBlanketTextureEverything.addActionListener(this);
		this.mnuTextures.add(this.mniBlanketTextureEverything);

		this.mniFindReplace.addActionListener(this);
		this.mnuTextures.add(this.mniFindReplace);

		this.mniLoadTextureAttributes.addActionListener(this);
		this.mnuTextures.add(this.mniLoadTextureAttributes);

		this.mniReloadTextures.addActionListener(this);
		this.mnuTextures.add(this.mniReloadTextures);

		this.add(this.mnuTextures);
	}

	@Override
	public void windowPopChange(boolean isInternalFrame) {
		this.mniPopper.setEnabled(isInternalFrame);
	}
}
