package greyvarEditor.ui.windows;

import java.awt.event.WindowAdapter;
import java.awt.event.WindowEvent;
import java.util.Vector;

import javax.swing.JFrame;
import javax.swing.JInternalFrame;
import javax.swing.event.InternalFrameEvent;
import javax.swing.event.InternalFrameListener;

import jwrCommonsJava.Configuration;
import jwrCommonsJava.Logger;

public abstract class PopableWindow extends JInternalFrame implements InternalFrameListener {
	public static interface Listener {
		public void windowPopChange(boolean isInternalFrame);
	}

	private final Vector<Listener> listeners = new Vector<Listener>();

	public PopableWindow() {
		this.addInternalFrameListener(this);
		this.setResizable(true);
		this.setClosable(true);
	}

	public void addListener(Listener l) {
		this.listeners.add(l);
	}

	@Override
	public void internalFrameActivated(InternalFrameEvent e) {
	}

	@Override
	public void internalFrameClosed(InternalFrameEvent e) {
		WindowMain.getInstance().fireWindowClosed(this);
	}

	@Override
	public void internalFrameClosing(InternalFrameEvent e) {
		// TODO Auto-generated method stub

	}

	@Override
	public void internalFrameDeactivated(InternalFrameEvent e) {
		// TODO Auto-generated method stub

	}

	@Override
	public void internalFrameDeiconified(InternalFrameEvent e) {
		// TODO Auto-generated method stub

	}

	@Override
	public void internalFrameIconified(InternalFrameEvent e) {
		// TODO Auto-generated method stub

	}

	@Override
	public void internalFrameOpened(InternalFrameEvent e) {
		WindowMain.getInstance().fireWindowOpened(this);
	}

	protected boolean isInternalFrame() {
		for (JInternalFrame fra : this.getDesktopPane().getAllFrames()) {
			if (fra.getTitle().equals(this.getTitle())) {
				return true;
			}
		}

		return false;
	}

	public void pop() {
		if (this.isInternalFrame()) {
			this.setVisible(false);
			this.setEnabled(false);

			JFrame frame = new JFrame();
			frame.setTitle(this.getTitle());
			frame.setContentPane(this.getContentPane());
			frame.setBounds(this.getBounds());
			frame.setLocationRelativeTo(null);
			frame.setJMenuBar(this.getJMenuBar());
			frame.setVisible(true);
			frame.addWindowListener(new WindowAdapter() {

				@Override
				public void windowClosing(WindowEvent e) {
					Logger.messageDebug("Closing");
					PopableWindow.this.setVisible(true);
					PopableWindow.this.setEnabled(true);
					PopableWindow.this.setBounds(Configuration.getBounds("windowEditorGrid", 100, 100, 320, 240));

					JFrame externalFrame = (JFrame) e.getWindow();

					PopableWindow.this.setContentPane(externalFrame.getContentPane());
					PopableWindow.this.setJMenuBar(externalFrame.getJMenuBar());

					for (Listener l : PopableWindow.this.listeners) {
						l.windowPopChange(true);
					}
				}
			});

			this.windowPopped();

			for (Listener l : this.listeners) {
				l.windowPopChange(false);
			}
		}
	}

	protected void windowPopped() {

	}
}
