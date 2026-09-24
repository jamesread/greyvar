package greyvarEditor.ui.windows;

import greyvarEditor.Tile;
import greyvarEditor.ui.components.ComponentTextureViewer;

import java.awt.GridBagConstraints;
import java.awt.GridBagLayout;
import java.awt.GridLayout;

import javax.swing.BorderFactory;
import javax.swing.JButton;
import javax.swing.JCheckBox;
import javax.swing.JFrame;
import javax.swing.JOptionPane;
import javax.swing.JPanel;

import jwrCommonsJava.Util;
import jwrCommonsJava.ui.JButtonWithAl;

public class WindowFindReplace extends JFrame {

	private final JPanel panFind = new JPanel();
	private final JPanel panRepl = new JPanel();

	private final JButton btnApply = new JButtonWithAl("Apply") {

		@Override
		public void click() {
			WindowFindReplace.this.doFindReplace();
		}
	};

	private final WindowEditorGrid windowEditorGrid;

	private final ComponentTextureViewer findTexture = new ComponentTextureViewer(true, null);
 
	private final ComponentTextureViewer replTexture = new ComponentTextureViewer(true, null);

	private final JCheckBox chkTraversable = new JCheckBox("Traversable");

	public WindowFindReplace(WindowEditorGrid windowEditorGrid) {
		this.windowEditorGrid = windowEditorGrid;

		this.setBounds(100, 100, 640, 480);
		this.setTitle("Find/Replace");

		this.setupComponents();

		this.setVisible(true);
	}

	private void doFindReplace() {
		int countReplacements = 0;

		for (Tile t : this.windowEditorGrid.gridEditor.getTiles()) {
			if (t.tex == this.findTexture.getTexture()) {
				t.tex = this.replTexture.getTexture();
				t.traversable = this.chkTraversable.isSelected();
				countReplacements++;
			}
		}

		JOptionPane.showMessageDialog(this, "Replaced: " + countReplacements + " tile(s).");
	}

	private void setupComponents() {
		this.setLayout(new GridBagLayout());
		GridBagConstraints gbc = Util.getNewGbc();

		this.panFind.add(this.findTexture);

		this.panFind.setBorder(BorderFactory.createTitledBorder("Find"));
		this.add(this.panFind, gbc);

		gbc.gridx++;
		this.panRepl.setBorder(BorderFactory.createTitledBorder("Replace"));
		this.panRepl.setLayout(new GridLayout(10, 1));
		this.panRepl.add(this.replTexture);
		this.panRepl.add(this.chkTraversable);
		this.add(this.panRepl, gbc);

		gbc.fill = 0;
		gbc.gridx = 1;
		gbc.anchor = GridBagConstraints.SOUTHEAST;
		gbc.gridy++;
		this.add(this.btnApply, gbc);
	}

}
