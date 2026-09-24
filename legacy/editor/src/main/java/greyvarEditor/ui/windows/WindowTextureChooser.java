package greyvarEditor.ui.windows;

import greyvarEditor.TextureCache;
import greyvarEditor.ui.windows.editors.grid.panels.Texture;
import greyvarEditor.utils.EditLayerMode;

import java.awt.BorderLayout;
import java.awt.Color;
import java.awt.Dimension;
import java.awt.GridLayout;
import java.awt.event.KeyEvent;
import java.awt.event.KeyListener;
import java.awt.event.MouseEvent;
import java.awt.event.MouseListener;
import java.util.Map;
import java.util.TreeSet;
import java.util.regex.Pattern;

import javax.swing.BorderFactory;
import javax.swing.ImageIcon;
import javax.swing.JFrame;
import javax.swing.JLabel;
import javax.swing.JPanel;
import javax.swing.JTextField;
import javax.swing.WindowConstants;

import jwrCommonsJava.Configuration;
import jwrCommonsJava.ui.JButtonWithAl;

public class WindowTextureChooser extends JFrame {
	private class JTextureButton extends JLabel implements MouseListener {
		private final Texture tex;

		public JTextureButton(Texture ttex) { 
			super(ttex.getId());

			this.addMouseListener(this);

			this.tex = ttex;
			this.setIcon(new ImageIcon(this.tex.image.getScaledInstance(64, 64, java.awt.Image.SCALE_SMOOTH)));
			this.setMaximumSize(new Dimension(64, 64));
			this.setBackground(Color.WHITE);
			this.setOpaque(true);
			this.setBorder(BorderFactory.createLineBorder(Color.WHITE));
			this.setHorizontalTextPosition(JLabel.CENTER);
			this.setVerticalTextPosition(JLabel.BOTTOM);
		}

		protected void click() {
			WindowTextureChooser.this.onClickedTexture(JTextureButton.this.tex);
		}

		@Override
		public void mouseClicked(MouseEvent e) {
			this.click();

			if (e.getClickCount() > 1) { 
				WindowTextureChooser.this.close();
				WindowTextureChooser.this.setVisible(false);
			}  
		}

		@Override
		public void mouseEntered(MouseEvent e) {
			this.setBorder(BorderFactory.createLineBorder(Color.BLUE));
		}

		@Override
		public void mouseExited(MouseEvent e) {
			this.setBorder(BorderFactory.createLineBorder(Color.WHITE));
		}

		@Override
		public void mousePressed(MouseEvent e) {
		}

		@Override
		public void mouseReleased(MouseEvent e) {
		}
	}

	public interface Listener {
		public void onTexChoose(Texture tex);
	}

	private class SearchInput extends JTextField implements KeyListener {
		public SearchInput() {
			this.addKeyListener(this);
		}

		@Override
		public void keyPressed(KeyEvent e) {
			if (e.getKeyCode() == KeyEvent.VK_ENTER) {
				WindowTextureChooser.this.updateSearch();
			}
		}

		@Override
		public void keyReleased(KeyEvent e) {
		}

		@Override
		public void keyTyped(KeyEvent e) {
		}
	}

	private final SearchInput searchInput = new SearchInput();

	private final Listener listener;
	
	private final TextureCache textureCache;

	private final JPanel panTextures = new JPanel();
	private final JLabel lblTextures = new JLabel("???");

	public WindowTextureChooser(Listener listener, EditLayerMode mode) {
		switch (mode) {
		case ENTITIES:
			textureCache = TextureCache.instanceEntities;
			break;
		case TILES:
			textureCache = TextureCache.instanceTiles;
			break;
		case FLUIDS: 
			textureCache = TextureCache.instanceFluids;
			break;
		default:
			textureCache = TextureCache.instanceTiles; 
		} 
		
		this.listener = listener;
		this.setupComponents();

		this.setTitle("Chooser");
		this.setDefaultCloseOperation(WindowConstants.DISPOSE_ON_CLOSE);
		this.setBounds(Configuration.getBounds("windowTextureChooser", 100, 100, 640, 480));
		this.setVisible(true);
	}

	@Override
	public void dispose() {
		close();
	}
	
	private void close() {
		Configuration.setBounds("windowTextureChooser", this.getBounds());
		this.setVisible(false);
	}

	public void onClickedTexture(Texture tex) {
		this.listener.onTexChoose(tex);
	}

	private void refreshTextureButtons() {
		this.refreshTextureButtons(null);
	}

	private void refreshTextureButtons(String keyword) { 
		this.panTextures.removeAll();
		Map<String, Texture> textures = textureCache.getAll();

		final TreeSet<String> sorted = new TreeSet<String>();
		sorted.addAll(textures.keySet());

		Pattern p = null;
		if ((keyword != null) && !keyword.isEmpty()) {
			p = Pattern.compile(keyword, Pattern.CASE_INSENSITIVE);
		}

		for (String tex : sorted) {
			if (p != null) {
				if (p.matcher(tex).find()) {
					this.panTextures.add(new JTextureButton(textures.get(tex)));
				}
			} else {
				this.panTextures.add(new JTextureButton(textures.get(tex)));
			}
		}

		this.lblTextures.setText(textures.size() + " texture(s)");
		this.panTextures.updateUI();
	}

	private void setupComponents() {
		this.setIconImage(WindowMain.getInstance().getIconImage());
		
		this.refreshTextureButtons();

		this.setLayout(new BorderLayout());
		this.panTextures.setLayout(new GridLayout(8, 8));
		this.panTextures.setBorder(BorderFactory.createLoweredBevelBorder());
		this.panTextures.setBackground(Color.DARK_GRAY);
		this.add(this.panTextures, BorderLayout.CENTER);

		JPanel panControls = new JPanel(new GridLayout(1, 0));

		panControls.add(this.searchInput);
		panControls.add(this.lblTextures);
		panControls.add(new JButtonWithAl("Refresh") {
			@Override 
			public void click() { 
				textureCache.loadTextures();
				WindowTextureChooser.this.refreshTextureButtons();
			}
		});

		this.add(panControls, BorderLayout.SOUTH);
	}

	private void updateSearch() {
		if (this.searchInput.getText().isEmpty()) {
			this.refreshTextureButtons();
		} else {
			this.refreshTextureButtons(this.searchInput.getText());
		}
	}
}
