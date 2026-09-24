package greyvarEditor.ui.components;

import java.awt.Dimension;
import java.awt.GridLayout;
import java.util.HashMap;

import javax.swing.JButton;
import javax.swing.JComponent;
import javax.swing.JInternalFrame;
import javax.swing.JPanel;

import greyvarEditor.ui.windows.WindowEditorGrid;
import greyvarEditor.ui.windows.WindowMain.InternalWindowListener;
import jwrCommonsJava.ui.JButtonWithAl;

public class ComponentWindowList extends JPanel implements InternalWindowListener {
	private class InternalWindowButton extends JButtonWithAl {
		private final JInternalFrame fra;

		public InternalWindowButton(JInternalFrame fra) { 
			super("");
			this.fra = fra;

			this.calculateButtonTitle();
			this.setMaximumSize(new Dimension(999, 100));
			this.setOpaque(true);
		}

		private void calculateButtonTitle() {
			String title = this.fra.getTitle();

			if ((title == null) || title.isEmpty()) {
				title = "(untitled)";
			}

			if (this.fra instanceof WindowEditorGrid) {
				title = "Grid: " + ((WindowEditorGrid) this.fra).getGridFile().getFilename();
			} else {
				title = "Window: " + title;
			}

			this.setText(title);
		}

		@Override
		public void click() {
			if ((this.fra instanceof WindowEditorGrid) && this.fra.isVisible()) {
				WindowEditorGrid weg = ((WindowEditorGrid) this.fra);
				this.setIcon(weg.getScreenshot());
			}

			this.fra.getDesktopPane().getDesktopManager().activateFrame(this.fra);
			this.fra.moveToFront();
		}
	}

	private final HashMap<JInternalFrame, JButton> winsToButtons = new HashMap<JInternalFrame, JButton>();

	public ComponentWindowList() {
		this.setLayout(new GridLayout(1, 99));
	}

	@Override
	public void windowClosed(JInternalFrame fra) {
		JComponent b = this.winsToButtons.get(fra);

		if ((b != null)) {
			this.remove(b);
		}
	}

	@Override
	public void windowOpened(JInternalFrame fra) {
		JButton b = new InternalWindowButton(fra);
		this.winsToButtons.put(fra, b);
		this.add(b);
	}
 
}
