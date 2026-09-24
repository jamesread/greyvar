package greyvarEditor.ui.windows;

import greyvarEditor.ui.components.ComponentWindowList;
import greyvarEditor.ui.menubars.MenuBarWindowMain;

import java.awt.BorderLayout;
import java.awt.Color;
import java.io.File;
import java.util.Collections;
import java.util.List;
import java.util.Vector;

import javax.imageio.ImageIO;
import javax.swing.JDesktopPane;
import javax.swing.JFrame;
import javax.swing.JInternalFrame;
import javax.swing.JLabel;
import javax.swing.JToolBar;
import javax.swing.WindowConstants;

import jwrCommonsJava.Configuration;
import jwrCommonsJava.Logger;

public class WindowMain extends JFrame {
	public interface InternalWindowListener {
		void windowClosed(JInternalFrame fra);

		void windowOpened(JInternalFrame fra);
	}

	private static final WindowMain instance = new WindowMain();

	public static WindowMain getInstance() {
		return WindowMain.instance;
	}

	public final JDesktopPane desktop = new JDesktopPane();

	private final JToolBar toolbar = new JToolBar("Window list");

	private final List<InternalWindowListener> listeners = Collections.synchronizedList(new Vector<InternalWindowListener>());

	private WindowMain() {
		this.setupComponents();

		this.setTitle("Greyvar editor");
		this.setBounds(Configuration.getBounds("windowMain", 100, 100, 320, 240));
		this.setDefaultCloseOperation(WindowConstants.DISPOSE_ON_CLOSE);
		this.setVisible(true);
	}

	public void addFrame(JInternalFrame wnd) {
		this.desktop.add(wnd);
	}

	public void addListener(InternalWindowListener listener) {
		this.listeners.add(listener);
	}

	@Override
	public void dispose() {
		this.setVisible(false);
		this.saveWindowPosition();
		Configuration.save();
		System.exit(0);
	}

	public void fireWindowClosed(JInternalFrame fra) {
		for (InternalWindowListener l : this.listeners) {
			l.windowClosed(fra);
		}
	}

	public void fireWindowOpened(JInternalFrame fra) {
		for (InternalWindowListener l : this.listeners) {
			l.windowOpened(fra);
		}
	}

	public void saveWindowPosition() {
		Logger.messageDebug("Saving window position");
		Configuration.setBounds("windowMain", this.getBounds());
	}

	private void setupComponents() {
		this.setLayout(new BorderLayout());
		
		try { 
			this.setIconImage(ImageIO.read(new File("res/logo.png")));
		} catch (Exception e) {} 

		ComponentWindowList winList = new ComponentWindowList();
		this.addListener(winList);

		this.toolbar.add(new JLabel("Windows list"));
		this.toolbar.add(winList);
		this.add(this.toolbar, BorderLayout.NORTH);

		this.desktop.setBackground(Color.DARK_GRAY);
		this.getContentPane().add(this.desktop, BorderLayout.CENTER);

		this.setJMenuBar(new MenuBarWindowMain());

	}

}
