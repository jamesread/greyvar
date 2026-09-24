package greyvarEditor.ui.windows.editors.grid.panels;

import greyvarEditor.Tile;
import greyvarEditor.ui.components.AllEnabledOrDisabledPanel;

import javax.swing.BorderFactory;
import javax.swing.JCheckBox;
import javax.swing.JPanel;

import jwrCommonsJava.ui.JCheckboxWithAl;

public class PanPhysics extends AllEnabledOrDisabledPanel {
	private final JCheckBox chkTraversable = new JCheckboxWithAl("Traversable") {
		@Override
		protected void onClicked() {
			PanPhysics.this.selected.traversable = this.isSelected();
		};
	};

	private Tile selected;

	public PanPhysics() {
		this.setBorder(BorderFactory.createTitledBorder("Physics"));
		this.setupComponents();

		this.chkTraversable.setEnabled(false);
	}

	private void setupComponents() {
		this.add(this.chkTraversable);
	}

	public void updateInterface(Tile t) {
		this.selected = t;

		this.chkTraversable.setEnabled(true);
		this.chkTraversable.setSelected(this.selected.traversable);
	}

}
