package greyvarEditor.utils;

import java.io.BufferedReader;
import java.io.File;
import java.io.FileNotFoundException;
import java.io.FileReader;
import java.io.FileWriter;
import java.io.IOException;
import java.util.Calendar;
import java.util.Vector;

import jwrCommonsJava.Logger;

public abstract class CsvFile { 
	protected final File f;
	 
	public CsvFile(File f) {
		this.f = f; 
	}
	
	public void load() {
		for (String line : this.getFileLines(this.f)) {
			if (line.startsWith("#")) {
				continue;
			}

			this.loadLine(line);
		}
	}
	
	protected abstract void loadLine(String line);
	 
	protected String getCsvLine(Object... obs) {
		String ret = "";

		for (Object o : obs) {
			if (o == null) {
				ret += "null,";
			} else {
				ret += o.toString() + ",";
			}
		}

		return ret + "\n";
	}
	
	protected Vector<String> getFileLines(File f) {
		final Vector<String> lines = new Vector<String>();

		try {
			BufferedReader br = new BufferedReader(new FileReader(f));

			while (true) {
				String line = br.readLine();

				if ((line == null) || line.isEmpty()) {
					break;
				} else {
					lines.add(line);
				}
			}

			br.close();
		} catch (FileNotFoundException e) {
			e.printStackTrace();
		} catch (IOException e) {
			e.printStackTrace();
		}

		return lines;
	}
	
	public String getFilename() {
		return this.f.getName();
	}
	
	public void save() {
		this.save(this.f);
	}

	public void save(File f) {
		try {
			FileWriter fw = new FileWriter(f);
			fw.append("# Saved at " + Calendar.getInstance().getTimeInMillis() + "\n");

			this.saveImpl(fw);

			fw.flush();
			fw.close();

			Logger.messageDebug("Saved file: " + f.getName());
		} catch (IOException e) {
			e.printStackTrace();
		}
	} 
	
	protected abstract void saveImpl(FileWriter fw);
}
