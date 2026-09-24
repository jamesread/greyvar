package greyvarEditor.ui.menubars;

import greyvarEditor.GameResources;
import greyvarEditor.dataModel.World;
import greyvarEditor.files.GridFile;
import greyvarEditor.files.GridFileLoader;
import greyvarEditor.files.WorldFileHandler;
import greyvarEditor.ui.windows.WindowEditorGrid;
import greyvarEditor.ui.windows.WindowEditorWorld;
import greyvarEditor.ui.windows.WindowMain;
import greyvarEditor.ui.windows.WindowOptions;

import java.awt.event.ActionEvent;
import java.awt.event.ActionListener;
import java.io.File;
import java.io.FileWriter;
import java.io.IOException;

import javax.swing.JFileChooser;
import javax.swing.JMenu;
import javax.swing.JMenuBar;
import javax.swing.JMenuItem;
import javax.swing.filechooser.FileFilter;

import jwrCommonsJava.Configuration;

public class MenuBarWindowMain extends JMenuBar implements ActionListener {
	private final JMenu mnuFile = new JMenu("File");
	private final JMenuItem mniOptions = new JMenuItem("Options");
	private final JMenuItem mniExit = new JMenuItem("Exit");

	private final JMenu mnuGrid = new JMenu("Grid");
	private final JMenuItem mniGridNew = new JMenuItem("New");
	private final JMenuItem mniGridOpen = new JMenuItem("Open");

	private final JMenu mnuWorld = new JMenu("World");
	private final JMenuItem mniWorldNew = new JMenuItem("New");
	private final JMenuItem mniWorldOpen = new JMenuItem("Open");

	public MenuBarWindowMain() {
		this.mniOptions.addActionListener(this);
		this.mnuFile.add(this.mniOptions);

		this.mnuFile.addSeparator();

		this.mniExit.addActionListener(this);
		this.mnuFile.add(this.mniExit);
		this.add(this.mnuFile);

		this.add(this.mnuWorld);

		this.mniWorldNew.addActionListener(this);
		this.mnuWorld.add(this.mniWorldNew);

		this.mniWorldOpen.addActionListener(this);
		this.mnuWorld.add(this.mniWorldOpen);
		
		this.add(this.mnuGrid);

		this.mniGridNew.addActionListener(this);
		this.mnuGrid.add(this.mniGridNew);

		this.mniGridOpen.addActionListener(this);
		this.mnuGrid.add(this.mniGridOpen); 
	}

	@Override
	public void actionPerformed(ActionEvent evt) {
		Object source = evt.getSource();

		if (source == this.mniGridNew) {
			this.newGrid();
		} else if (source == this.mniWorldNew) {
			this.newWorld();
		} else if (source == this.mniWorldOpen) {
			this.loadWorld();
		} else if (source == this.mniGridOpen) {
			this.loadGrid();
		} else if (source == this.mniExit) {
			WindowMain.getInstance().saveWindowPosition();
			Configuration.save();
			System.exit(0);
		} else if (source == this.mniOptions) {
			WindowMain.getInstance().addFrame(new WindowOptions());
		} else {
			System.out.println("Unhandled action - window main");
		}
	}

	void loadGrid() {
		JFileChooser jfc = new JFileChooser();
		jfc.setBounds(Configuration.getBounds("loadGridJfc", 100, 100, 640, 480));
		jfc.setCurrentDirectory(new File(Configuration.getS("grids.path")));
		jfc.setDialogTitle("Load grid");
		jfc.setFileFilter(new FileFilter() {

			@Override
			public boolean accept(File f) {
				return f.isDirectory() || f.getName().endsWith("grid") || f.getName().endsWith("csv");
			}

			@Override
			public String getDescription() { 
				return "Grid files (*.grid, *.csv)";
			}
		});

		jfc.setCurrentDirectory(GameResources.dirServerData);

		if (jfc.showOpenDialog(this) == JFileChooser.APPROVE_OPTION) {
			Configuration.set("grids.path", jfc.getCurrentDirectory().getPath());
			this.loadGrid(jfc.getSelectedFile());
		}

		Configuration.setBounds("loadGridJfc", jfc.getBounds());
	}
	
	void loadGrid(File file) {
		loadGrid(GridFileLoader.load(file));  
	}

	void loadGrid(GridFile file) {  
		WindowMain.getInstance().addFrame(new WindowEditorGrid(file));
	}

	void loadWorld() {
		JFileChooser jfc = new JFileChooser();
		jfc.setDialogTitle("Load world");
		jfc.setCurrentDirectory(GameResources.dirServerDataWorlds);
		jfc.setFileFilter(new FileFilter() {
			@Override
			public String getDescription() {
				return "World Files";
			}
			
			@Override
			public boolean accept(File f) {
				if (f.isFile()) { 
					return f.getName().equals("world.yml");	
				} else {
					return true;
				} 
			}
		});

		if (jfc.showOpenDialog(this) == JFileChooser.APPROVE_OPTION) {
			this.loadWorld(WorldFileHandler.load((jfc.getSelectedFile())));
		}   
	}  

	void loadWorld(World wf) { 
		WindowMain.getInstance().addFrame(new WindowEditorWorld(wf));
	}
 
	void newGrid() {
		JFileChooser jfc = new JFileChooser();
		jfc.setDialogTitle("New grid");
		jfc.setCurrentDirectory(GameResources.dirServerData);
		jfc.setBounds(Configuration.getBounds("newGrid", 100, 100, 320, 240));

		if (jfc.showSaveDialog(this) == JFileChooser.APPROVE_OPTION) {
			File newFile = jfc.getSelectedFile();

			try {
				newFile.createNewFile();

				FileWriter fw = new FileWriter(newFile);
				fw.write("---");
				fw.close();
			} catch (IOException e) {
				e.printStackTrace();
				return;
			}

			this.loadGrid(jfc.getSelectedFile());
		}

		Configuration.setBounds("newGrid", jfc.getBounds());
	}

	void newWorld() {
		JFileChooser jfc = new JFileChooser();
		jfc.setDialogTitle("New world");
		jfc.setCurrentDirectory(GameResources.dirServerData);
		jfc.setBounds(Configuration.getBounds("newWorld", 100, 100, 320, 240));

		if (jfc.showSaveDialog(this) == JFileChooser.APPROVE_OPTION) {
			File newFile = jfc.getSelectedFile();

			try {
				newFile.createNewFile();
			} catch (IOException e) {
				e.printStackTrace();
				return;
			} 

			this.loadGrid(GridFileLoader.load(jfc.getSelectedFile()));
		}

		Configuration.setBounds("newWorld", this.getBounds());
	}

}
