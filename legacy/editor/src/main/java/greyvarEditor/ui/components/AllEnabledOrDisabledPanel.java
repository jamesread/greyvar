package greyvarEditor.ui.components;

import java.awt.Component;

import javax.swing.JPanel;

public class AllEnabledOrDisabledPanel extends JPanel {
	@Override
	public void setEnabled(boolean enabled) { 
		super.setEnabled(enabled);
		
		for (Component comp : this.getComponents()) {
			comp.setEnabled(enabled); 
		}
	}
}
