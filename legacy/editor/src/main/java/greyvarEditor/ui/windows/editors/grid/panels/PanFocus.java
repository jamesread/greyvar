package greyvarEditor.ui.windows.editors.grid.panels;

import greyvarEditor.Tile;
import greyvarEditor.dataModel.EntityInstance;
import greyvarEditor.ui.components.AllEnabledOrDisabledPanel;
import greyvarEditor.ui.components.ComponentTextureViewer;
import greyvarEditor.utils.EditLayerMode;

import java.awt.GridLayout;

import javax.swing.BorderFactory;
import javax.swing.JLabel;
import javax.swing.JPanel;

public class PanFocus extends AllEnabledOrDisabledPanel {
	public PanFocus() {
		this.setBorder(BorderFactory.createTitledBorder("Focus (red box)"));
	}

	public void setFocus(int x, int y, Tile t, EntityInstance e) {
		this.removeAll(); 
		this.setLayout(new GridLayout(1, 0));
		this.add(new JLabel("Cell: " + x + ":" + y));
		this.add(new ComponentTextureViewer(false, t.tex, EditLayerMode.ENTITIES));   
		this.doLayout(); 
		this.repaint(); 
	}

}
