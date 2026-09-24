package greyvarEditor.ui.components;

import greyvarEditor.ui.windows.WindowTextureChooser;
import greyvarEditor.ui.windows.editors.grid.panels.Texture;
import greyvarEditor.utils.EditLayerMode;

import java.awt.BorderLayout;
import java.awt.Color;
import java.awt.Dimension;

import javax.swing.BorderFactory;
import javax.swing.ImageIcon;
import javax.swing.JButton;
import javax.swing.JLabel;
import javax.swing.JPanel;

import jwrCommonsJava.ui.JButtonWithAl;
 
public class ComponentTextureViewer extends AllEnabledOrDisabledPanel implements WindowTextureChooser.Listener {
	private Texture tex = null;
	private final JLabel lblTex = new JLabel(); 

	private final JButton btnChoose = new JButtonWithAl("Choose") {
 
		@Override
		public void click() {
			new WindowTextureChooser(ComponentTextureViewer.this, mode);
		}
	};
 
	public ComponentTextureViewer() {
		this(false, null);
	}
	
	public ComponentTextureViewer(boolean chooserButton, Texture t) {
		this(chooserButton, t, EditLayerMode.TILES);  
	}
	
	private EditLayerMode mode;

	public ComponentTextureViewer(boolean chooserButton, Texture t, EditLayerMode mode) {
		this.mode = mode;
		
		this.lblTex.setPreferredSize(new Dimension(200, 200));
		this.lblTex.setBackground(Color.GRAY);
		this.lblTex.setBorder(BorderFactory.createLoweredBevelBorder());
		 
		this.unsetTex();

		this.setLayout(new BorderLayout());
		this.add(this.lblTex, BorderLayout.CENTER);

		if (chooserButton) {
			this.add(this.btnChoose, BorderLayout.EAST);
		}
		
		this.setTex(t);
	}

	public ComponentTextureViewer(Texture t) {
		this();
		this.setTex(t);
	}

	public Texture getTexture() {
		return this.tex;
	}

	@Override
	public void onTexChoose(Texture tex) {
		this.setTex(tex);
	}

	public void setTex(Texture tex) {
		if (tex == null) {
			this.unsetTex();
			return; 
		}
		
		this.tex = tex;
		this.lblTex.setIcon(new ImageIcon(tex.image));
		this.lblTex.setText(tex.getId());
		this.lblTex.repaint();
	}

	public void unsetTex() {
		this.tex = null;
		this.lblTex.setIcon(null);
		this.lblTex.setText("No texture");
		this.lblTex.revalidate();  
		this.lblTex.repaint(); 
	};
}
