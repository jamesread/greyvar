package greyvarEditor.ui.windows;

import java.awt.BorderLayout;

import javax.swing.JDialog;
import javax.swing.JProgressBar;

public class WindowProgress extends JDialog {
	private final JProgressBar proMain = new JProgressBar(0, 100);

	public WindowProgress() { 
		this.setSize(320, 120);
		this.setLocationRelativeTo(null);
		this.setTitle("Progress");

		this.setupComponents();
	}

	public void setProgress(int progress) {
		this.proMain.setValue(progress);

		if (progress >= 100) {
			this.setVisible(false);
		}
	}

	private void setupComponents() {
		this.setLayout(new BorderLayout());
		this.add(this.proMain);
	}
}
