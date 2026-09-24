package greyvarEditor.ui.windows;

import greyvarEditor.Tile;
import greyvarEditor.dataModel.EntityInstance;
import greyvarEditor.files.GridFile;
import greyvarEditor.ui.components.ComponentGridEditor;
import greyvarEditor.ui.menubars.MenuBarEditorGrid;
import greyvarEditor.ui.windows.editors.grid.panels.PanAppearance;
import greyvarEditor.ui.windows.editors.grid.panels.PanEntity;
import greyvarEditor.ui.windows.editors.grid.panels.PanFocus;
import greyvarEditor.ui.windows.editors.grid.panels.PanPhysics;
import greyvarEditor.ui.windows.editors.grid.panels.PanTransport;
import greyvarEditor.ui.windows.editors.grid.panels.Texture;
import greyvarEditor.utils.EditLayerMode;

import java.awt.BorderLayout;
import java.awt.GridBagConstraints;
import java.awt.GridBagLayout;
import java.awt.Insets;

import javax.swing.Icon;
import javax.swing.ImageIcon;
import javax.swing.JOptionPane;
import javax.swing.JPanel;
import javax.swing.JScrollPane;
import javax.swing.JSplitPane;
import javax.swing.WindowConstants;
import javax.swing.event.InternalFrameAdapter;
import javax.swing.event.InternalFrameEvent;

import jwrCommonsJava.Configuration;
import jwrCommonsJava.Logger;

public class WindowEditorGrid extends PopableWindow {

	public final ComponentGridEditor gridEditor; 
	
	private final MenuBarEditorGrid mnu = new MenuBarEditorGrid(this);
 
	private JSplitPane sp;

	private final PanFocus panFocus = new PanFocus();
	public final PanAppearance panAppearance = new PanAppearance();
	private final PanPhysics panPhysics = new PanPhysics(); 
	private final PanTransport panTransport = new PanTransport();
	public final PanEntity panEntity = new PanEntity();  
 
	private final GridFile gf;

	public WindowEditorGrid(GridFile gf) {
		this.setDefaultCloseOperation(WindowConstants.DO_NOTHING_ON_CLOSE);
		this.addInternalFrameListener(new InternalFrameAdapter() {
			@Override
			public void internalFrameClosing(InternalFrameEvent e) {
				WindowEditorGrid.this.windowClose();
			}
		});
 
		this.gf = gf;
		this.gridEditor = new ComponentGridEditor(this);
		this.gridEditor.setGridFile(gf);
		 
		this.setEditMode(EditLayerMode.TILES); 

		this.setVisible(true);
		this.setTitle("Grid Editor - " + gf.getFilename());
		this.setBounds(Configuration.getBounds("windowEditorGrid", 100, 100, 320, 240));
		this.setResizable(true);
		this.setMaximizable(true);
		this.setClosable(true);
		this.setIconifiable(true);

		this.requestFocus();

		this.setupComponents();
	}

	public void blanketTextureEverything() {
		Texture t = this.panAppearance.getCurrentTexture();

		if (t == null) {
			JOptionPane.showMessageDialog(this, "Select a texture first!");
		} else {
			this.gridEditor.blanketTextureEverything(t);
		}
	}
	
	public void setEditMode(EditLayerMode mode) {
		this.gridEditor.setEditMode(mode); 
		
		switch (mode) { 
		case TILES: 
			this.panAppearance.setEnabled(true);
			this.panEntity.setEnabled(false);
			this.panPhysics.setEnabled(true);
			this.panTransport.setEnabled(true);
			break;
		case ENTITIES: 
			this.panAppearance.setEnabled(false);
			this.panEntity.setEnabled(true);
			this.panPhysics.setEnabled(false); 
			this.panTransport.setEnabled(false);
			break; 
		case FLUIDS:
			this.panAppearance.setEnabled(false); 
			this.panEntity.setEnabled(false); 
			this.panPhysics.setEnabled(false); 
			this.panTransport.setEnabled(false);
			break; 
		} 
		 
		this.mnu.toggleLayerButtons();
	}

	public GridFile getGridFile() {
		return this.gf;
	}

	public Icon getScreenshot() {
		return new ImageIcon(this.gridEditor.getScreenshot());
	}

	public void onCellFocus(int x, int y, Tile t, EntityInstance e) {
		this.panFocus.setFocus(x, y, t, e);
		this.panPhysics.updateInterface(t);
		this.panTransport.updateSelectedTile(x, y, t);
	} 

	public void onCellSelected(int currx, int curry, Tile tile, EntityInstance entity) {
		this.panAppearance.setCurrentTexture(tile.tex);
		this.panEntity.setCurrentEntity(entity); 
		this.panFocus.setFocus(currx, curry, tile, entity);
	}

	public void reload() { 
		this.gf.load();
		this.gridEditor.setGridFile(this.gf);
	}

	private void setupComponents() {
		this.setJMenuBar(this.mnu);

		JPanel panControls = new JPanel(new GridBagLayout());
		GridBagConstraints gbc = new GridBagConstraints(0, 0, 1, 1, 1, 1, GridBagConstraints.CENTER, 1, new Insets(6, 6, 6, 6), 0, 0);

		gbc.weighty = 0;
		panControls.add(this.panFocus, gbc);

		gbc.gridy++;
		panControls.add(this.panAppearance, gbc);
		
		gbc.gridy++; 
		panControls.add(this.panEntity, gbc);

		gbc.gridy++;
		panControls.add(this.panPhysics, gbc);

		gbc.weighty = 1;
		gbc.gridy++;
		panControls.add(this.panTransport, gbc);

		this.sp = new JSplitPane(JSplitPane.HORIZONTAL_SPLIT, new JScrollPane(this.gridEditor), panControls);
		this.sp.setDividerLocation(Configuration.getI("windowEditorGrid.sidebarPositionFromLeft", 300));
		this.add(this.sp, BorderLayout.CENTER);
	}

	public void windowClose() {
		int result = JOptionPane.showOptionDialog(this, "Close?", "Close?", JOptionPane.YES_NO_OPTION, JOptionPane.QUESTION_MESSAGE, null, null, JOptionPane.NO_OPTION);

		if (result == JOptionPane.YES_OPTION) {
			super.internalFrameClosed(null);
			this.setVisible(false);
			this.dispose();
		}

		Logger.messageDebug("Saving editor bounds");
		Configuration.setBounds("windowEditorGrid", this.getBounds());
		Configuration.set("windowEditorGrid.sidebarPositionFromLeft", this.sp.getDividerLocation());
	}

}
