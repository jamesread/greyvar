/// <reference types="@mapeditor/tiled-api" />

/**
 * Template Palette
 *
 * Grid palette of .tx object templates with tile previews (first animation
 * frame when present). Selecting a cell loads the template and activates
 * Insert Template for placement on object layers.
 *
 * Never call Dialog.clear() from inside a child widget signal handler — that
 * crashes Qt. Filtering toggles widget visibility; full rebuilds are deferred.
 */

const SESSION_DIR = "templatePalette.directory";
const SESSION_FILTER = "templatePalette.filter";
const PREVIEW_SCALE = 4;
const PALETTE_COLUMNS = 4;
const GID_FLIP_MASK = 0x1fffffff;

/** @type {Dialog | null} */
let paletteDialog = null;

/** @type {TemplateEntry[]} */
let templates = [];

/** @type {string} */
let selectedPath = "";

/** @type {string} */
let pendingFilter = "";

/** @type {Qt.QLineEdit | null} */
let filterEdit = null;
/** @type {Qt.QLabel | null} */
let scopeLabel = null;
/** @type {Qt.QLabel | null} */
let statusLabel = null;
/** @type {Qt.QPushButton | null} */
let projectBtn = null;

/**
 * @typedef {object} PaletteCell
 * @property {TemplateEntry} entry
 * @property {ImageWidget} imageWidget
 * @property {Qt.QPushButton} button
 */

/** @type {PaletteCell[]} */
let paletteCells = [];

/** @type {Object.<string, Tileset>} */
let tilesetCache = {};

/**
 * @typedef {object} TemplateEntry
 * @property {string} path
 * @property {string} name
 * @property {string} className
 * @property {string} label
 * @property {string} relative
 * @property {Image} previewImage
 */

/**
 * @param {() => void} fn
 */
function afterEvent(fn) {
	if (typeof Qt !== "undefined" && typeof Qt.callLater === "function") {
		Qt.callLater(fn);
		return;
	}
	tiled.warn(
		"Template Palette: Qt.callLater missing; close and reopen the palette instead of Refresh.",
		function () {}
	);
}

/**
 * @param {string} attrs
 * @param {string} name
 * @returns {string}
 */
function xmlAttr(attrs, name) {
	const match = attrs.match(new RegExp("\\b" + name + '="([^"]*)"'));
	return match ? match[1] : "";
}

/**
 * @param {string} path
 * @returns {{ name: string, className: string, firstGid: number, tilesetSource: string, gid: number }}
 */
function parseTemplate(path) {
	const fallback = FileInfo.completeBaseName(path);
	const result = {
		name: fallback,
		className: "",
		firstGid: 1,
		tilesetSource: "",
		gid: 0,
	};

	try {
		const file = new TextFile(path, TextFile.ReadOnly);
		const text = file.readAll();
		file.close();

		const tilesetTag = text.match(/<tileset\b([^>]*)\/?>/);
		if (tilesetTag) {
			result.firstGid = parseInt(xmlAttr(tilesetTag[1], "firstgid"), 10) || 1;
			result.tilesetSource = xmlAttr(tilesetTag[1], "source");
		}

		const objectTag = text.match(/<object\b([^>]*)\/?>/);
		if (objectTag) {
			const attrs = objectTag[1];
			result.className =
				xmlAttr(attrs, "class") || xmlAttr(attrs, "type") || "";
			const name = xmlAttr(attrs, "name");
			if (name) result.name = name;
			result.gid = parseInt(xmlAttr(attrs, "gid"), 10) || 0;
		}
	} catch (e) {
		tiled.log("Template Palette: could not parse " + path + ": " + e);
	}

	return result;
}

/**
 * @param {string} root
 * @param {string[]} out
 */
function collectTxFiles(root, out) {
	if (!root || !File.exists(root)) return;

	const dirs = File.directoryEntries(root, File.Dirs | File.NoDotAndDotDot);
	for (let i = 0; i < dirs.length; ++i)
		collectTxFiles(FileInfo.joinPaths(root, dirs[i]), out);

	const files = File.directoryEntries(root, File.Files | File.NoDotAndDotDot);
	for (let i = 0; i < files.length; ++i) {
		const name = files[i];
		if (FileInfo.suffix(name).toLowerCase() === "tx")
			out.push(FileInfo.joinPaths(root, name));
	}
}

/**
 * @returns {string[]}
 */
function scanRoots() {
	const custom = String(tiled.session.get(SESSION_DIR, "") || "");
	if (custom && File.exists(custom)) return [FileInfo.cleanPath(custom)];

	const roots = [];
	const project = tiled.project;
	if (project && project.fileName && project.folders && project.folders.length) {
		const projectDir = FileInfo.path(project.fileName);
		for (let i = 0; i < project.folders.length; ++i) {
			const folder = project.folders[i];
			const abs = FileInfo.isAbsolutePath(folder)
				? folder
				: FileInfo.joinPaths(projectDir, folder);
			roots.push(FileInfo.cleanPath(abs));
		}
	}

	if (!roots.length && tiled.activeAsset && tiled.activeAsset.fileName)
		roots.push(FileInfo.path(tiled.activeAsset.fileName));

	return roots;
}

/**
 * @param {string} path
 * @param {string[]} roots
 * @returns {string}
 */
function displayRelative(path, roots) {
	for (let i = 0; i < roots.length; ++i) {
		const root = roots[i];
		if (path === root || path.indexOf(root + "/") === 0)
			return FileInfo.relativePath(root, path);
	}
	return FileInfo.fileName(path);
}

/**
 * @param {string} tsxPath
 * @returns {Tileset | null}
 */
function loadTileset(tsxPath) {
	if (!tsxPath || !File.exists(tsxPath)) return null;
	if (tilesetCache[tsxPath]) return tilesetCache[tsxPath];

	try {
		const format =
			tiled.tilesetFormatForFile(tsxPath) || tiled.tilesetFormat("tsx");
		if (!format) return null;
		const tileset = format.read(tsxPath);
		tilesetCache[tsxPath] = tileset;
		return tileset;
	} catch (e) {
		tiled.log("Template Palette: failed to load tileset " + tsxPath + ": " + e);
		return null;
	}
}

/**
 * @returns {Image}
 */
function makePlaceholderImage() {
	const size = 15 * PREVIEW_SCALE;
	const image = new Image(size, size, Image.Format_ARGB32);
	image.fill("#5c6770");
	const border = tiled.color("#2d6a4f");
	for (let x = 0; x < size; ++x) {
		image.setPixelColor(x, 0, border);
		image.setPixelColor(x, size - 1, border);
	}
	for (let y = 0; y < size; ++y) {
		image.setPixelColor(0, y, border);
		image.setPixelColor(size - 1, y, border);
	}
	return image;
}

/**
 * First-frame tile image for a template, scaled for the palette.
 * @param {string} txPath
 * @param {{ firstGid: number, tilesetSource: string, gid: number }} meta
 * @returns {Image}
 */
function renderPreview(txPath, meta) {
	if (!meta.gid || !meta.tilesetSource) return makePlaceholderImage();

	const tsxPath = FileInfo.cleanPath(
		FileInfo.joinPaths(FileInfo.path(txPath), meta.tilesetSource)
	);
	const tileset = loadTileset(tsxPath);
	if (!tileset) return makePlaceholderImage();

	const localId = (meta.gid & GID_FLIP_MASK) - meta.firstGid;
	if (localId < 0) return makePlaceholderImage();

	let tile = tileset.findTile(localId);
	if (!tile) return makePlaceholderImage();

	if (tile.frames && tile.frames.length > 0) {
		const frameTile = tileset.findTile(tile.frames[0].tileId);
		if (frameTile) tile = frameTile;
	}

	try {
		const image = tile.image.copy(tile.imageRect);
		if (!image || image.width < 1 || image.height < 1)
			return makePlaceholderImage();

		return image.scaled(
			Math.max(1, image.width * PREVIEW_SCALE),
			Math.max(1, image.height * PREVIEW_SCALE),
			Image.IgnoreAspectRatio,
			Image.FastTransformation
		);
	} catch (e) {
		tiled.log("Template Palette: preview failed for " + txPath + ": " + e);
		return makePlaceholderImage();
	}
}

/**
 * @returns {string[]}
 */
function refreshTemplateList() {
	tilesetCache = {};
	const roots = scanRoots();
	const paths = [];
	for (let i = 0; i < roots.length; ++i) collectTxFiles(roots[i], paths);
	paths.sort();

	templates = [];
	for (let i = 0; i < paths.length; ++i) {
		const path = paths[i];
		const meta = parseTemplate(path);
		const relative = displayRelative(path, roots);

		let label = meta.name;
		if (meta.className && meta.className !== meta.name)
			label += " [" + meta.className + "]";

		templates.push({
			path: path,
			name: meta.name,
			className: meta.className,
			label: label,
			relative: relative,
			previewImage: renderPreview(path, meta),
		});
	}

	return roots;
}

/**
 * @param {string} filter
 * @returns {TemplateEntry[]}
 */
function filteredTemplates(filter) {
	const q = String(filter || "")
		.trim()
		.toLowerCase();
	if (!q) return templates.slice();

	return templates.filter(function (t) {
		return (
			t.label.toLowerCase().indexOf(q) !== -1 ||
			t.path.toLowerCase().indexOf(q) !== -1 ||
			t.className.toLowerCase().indexOf(q) !== -1 ||
			t.name.toLowerCase().indexOf(q) !== -1
		);
	});
}

function activateInsertTemplateTool() {
	const editor = tiled.mapEditor;
	if (editor && typeof editor.tool === "function") {
		const tool = editor.tool("CreateTemplateTool");
		if (tool) {
			editor.selectedTool = tool;
			return true;
		}
	}

	try {
		tiled.trigger("CreateTemplateTool");
		return true;
	} catch (e) {
		tiled.warn(
			"Template Palette: could not activate Insert Template tool. " +
				"Select it manually (shortcut V) after choosing a template.",
			function () {}
		);
		return false;
	}
}

/**
 * @param {boolean} selected
 * @returns {string}
 */
function cellButtonStyle(selected) {
	const border = selected ? "#2d6a4f" : "#888888";
	const borderWidth = selected ? "2px" : "1px";
	const bg = selected ? "#d8f3dc" : "#f0f0f0";
	return (
		"QPushButton {" +
		"  min-width: 72px;" +
		"  max-width: 110px;" +
		"  padding: 4px 6px;" +
		"  background-color: " +
		bg +
		";" +
		"  border: " +
		borderWidth +
		" solid " +
		border +
		";" +
		"  border-radius: 3px;" +
		"}" +
		"QPushButton:hover { border-color: #2d6a4f; }" +
		"QPushButton:checked {" +
		"  border: 2px solid #2d6a4f;" +
		"  background-color: #d8f3dc;" +
		"}"
	);
}

/**
 * @param {string} path
 */
function selectTemplate(path) {
	if (!path || !File.exists(path)) {
		tiled.alert("Template file not found:\n" + path);
		return;
	}

	selectedPath = path;
	tiled.open(path);
	activateInsertTemplateTool();
	tiled.log("Template Palette: ready to place " + path);

	syncPaletteSelection();

	if (statusLabel) {
		statusLabel.text =
			"Ready to place: " +
			FileInfo.fileName(path) +
			" — click the map (object layer)";
	}
}

function syncPaletteSelection() {
	for (let i = 0; i < paletteCells.length; ++i) {
		const cell = paletteCells[i];
		const selected = cell.entry.path === selectedPath;
		cell.button.checked = selected;
		cell.button.styleSheet = cellButtonStyle(selected);
		cell.button.text = selected ? "✓ " + cell.entry.name : cell.entry.name;
	}
}

/**
 * Toggle palette cell visibility from the filter — no Dialog.clear().
 * @param {string} [filterText]
 */
function applyFilterInPlace(filterText) {
	if (filterText === undefined)
		filterText = filterEdit ? filterEdit.text : pendingFilter;
	pendingFilter = String(filterText || "");
	tiled.session.set(SESSION_FILTER, pendingFilter);

	const visible = filteredTemplates(pendingFilter);
	/** @type {Object.<string, boolean>} */
	const allowed = {};
	for (let i = 0; i < visible.length; ++i) allowed[visible[i].path] = true;

	let shown = 0;
	for (let i = 0; i < paletteCells.length; ++i) {
		const show = !!allowed[paletteCells[i].entry.path];
		paletteCells[i].imageWidget.visible = show;
		paletteCells[i].button.visible = show;
		if (show) shown++;
	}

	if (!statusLabel) return;
	if (!templates.length) {
		statusLabel.text =
			"No .tx files found. Add templates to the project or choose a folder.";
	} else if (!shown) {
		statusLabel.text = "No templates match the filter.";
	} else if (selectedPath && allowed[selectedPath]) {
		statusLabel.text = "Selected: " + FileInfo.fileName(selectedPath);
	} else {
		statusLabel.text =
			shown + " template(s). Click a name under a preview to place.";
	}
}

/**
 * @param {string[]} roots
 */
function updateScopeLabel(roots) {
	if (!scopeLabel) return;
	const customDir = String(tiled.session.get(SESSION_DIR, "") || "");
	scopeLabel.text = customDir
		? "Folder: " + customDir
		: roots.length
			? "Scanning project folders (" + templates.length + " templates)"
			: "No project folders — choose a folder to scan";
	if (projectBtn) projectBtn.enabled = !!customDir;
}

/**
 * Build one strip of previews and the matching name buttons under them.
 * @param {Dialog} dialog
 * @param {TemplateEntry[]} entries
 */
function addPaletteStrip(dialog, entries) {
	dialog.addNewRow();
	for (let i = 0; i < entries.length; ++i) {
		const entry = entries[i];
		const imageWidget = dialog.addImage("", entry.previewImage, entry.path);
		// button added in second pass so images share a row, buttons share the next
		paletteCells.push({
			entry: entry,
			imageWidget: imageWidget,
			button: /** @type {Qt.QPushButton} */ (/** @type {unknown} */ (null)),
		});
	}

	dialog.addNewRow();
	const start = paletteCells.length - entries.length;
	for (let i = 0; i < entries.length; ++i) {
		(function (cell) {
			const btn = dialog.addButton(
				cell.entry.path === selectedPath
					? "✓ " + cell.entry.name
					: cell.entry.name
			);
			btn.checkable = true;
			btn.checked = cell.entry.path === selectedPath;
			btn.toolTip =
				cell.entry.path +
				(cell.entry.className ? "\nclass: " + cell.entry.className : "");
			btn.styleSheet = cellButtonStyle(cell.entry.path === selectedPath);
			btn.clicked.connect(function () {
				selectTemplate(cell.entry.path);
			});
			cell.button = btn;
		})(paletteCells[start + i]);
	}
}

/**
 * Full dialog rebuild. Must not run inside a child widget signal handler.
 * @param {string} [filterText]
 */
function rebuildPaletteNow(filterText) {
	if (!paletteDialog) return;

	pendingFilter =
		filterText !== undefined
			? String(filterText || "")
			: filterEdit
				? filterEdit.text
				: String(tiled.session.get(SESSION_FILTER, "") || "");

	filterEdit = null;
	scopeLabel = null;
	statusLabel = null;
	projectBtn = null;
	paletteCells = [];

	const dialog = paletteDialog;
	dialog.clear();

	const roots = refreshTemplateList();
	const customDir = String(tiled.session.get(SESSION_DIR, "") || "");

	dialog.newRowMode = Dialog.ManualRows;

	dialog.addHeading("Template Palette", true);
	dialog.addLabel(
		"Click a template name under its preview, then click an object layer to place it."
	);

	dialog.addNewRow();
	scopeLabel = dialog.addLabel("");
	updateScopeLabel(roots);

	dialog.addNewRow();
	filterEdit = dialog.addTextInput("Filter:", pendingFilter);
	filterEdit.placeholderText = "name, class, path…";

	const applyBtn = dialog.addButton("Apply Filter");
	applyBtn.toolTip = "Filter visible templates (also runs when Filter loses focus)";

	dialog.addNewRow();
	const refreshBtn = dialog.addButton("Refresh");
	refreshBtn.toolTip = "Rescan for .tx files and rebuild previews";

	const folderBtn = dialog.addButton("Choose Folder…");
	folderBtn.toolTip = "Scan a specific directory for .tx templates";

	projectBtn = dialog.addButton("Use Project Folders");
	projectBtn.enabled = !!customDir;
	projectBtn.toolTip = "Clear folder override and scan project folders again";

	dialog.addNewRow();
	dialog.addHeading("Templates", true);

	if (!templates.length) {
		dialog.addNewRow();
		dialog.addLabel("No .tx templates found.");
	} else {
		for (let i = 0; i < templates.length; i += PALETTE_COLUMNS) {
			addPaletteStrip(dialog, templates.slice(i, i + PALETTE_COLUMNS));
		}
	}

	dialog.addNewRow();
	statusLabel = dialog.addLabel("");

	filterEdit.editingFinished.connect(function () {
		applyFilterInPlace(filterEdit.text);
	});

	applyBtn.clicked.connect(function () {
		applyFilterInPlace(filterEdit.text);
	});

	refreshBtn.clicked.connect(function () {
		const text = filterEdit.text;
		afterEvent(function () {
			rebuildPaletteNow(text);
		});
	});

	folderBtn.clicked.connect(function () {
		afterEvent(function () {
			chooseDirectory();
		});
	});

	projectBtn.clicked.connect(function () {
		const text = filterEdit ? filterEdit.text : pendingFilter;
		afterEvent(function () {
			tiled.session.set(SESSION_DIR, "");
			rebuildPaletteNow(text);
		});
	});

	applyFilterInPlace(pendingFilter);
}

function chooseDirectory() {
	const current = String(tiled.session.get(SESSION_DIR, "") || "");
	const start =
		current ||
		(tiled.project && tiled.project.fileName
			? FileInfo.path(tiled.project.fileName)
			: "");
	const dir = tiled.promptDirectory(start, "Template search folder");
	if (!dir) return;

	tiled.session.set(SESSION_DIR, dir);
	rebuildPaletteNow(filterEdit ? filterEdit.text : pendingFilter);
}

function showPalette() {
	if (!paletteDialog) {
		paletteDialog = new Dialog("Template Palette");
		paletteDialog.minimumWidth = 420;
		paletteDialog.finished.connect(function () {
			paletteDialog = null;
			filterEdit = null;
			scopeLabel = null;
			statusLabel = null;
			projectBtn = null;
			paletteCells = [];
		});
	}

	pendingFilter = String(tiled.session.get(SESSION_FILTER, "") || "");
	rebuildPaletteNow(pendingFilter);
	paletteDialog.show();
}

const openAction = tiled.registerAction("TemplatePalette", function () {
	showPalette();
});
openAction.text = "Template Palette";
openAction.icon = "template-palette.svg";
openAction.shortcut = "Ctrl+Shift+T";

tiled.extendMenu("View", [{ action: "TemplatePalette" }]);

tiled.log("Template Palette extension loaded (View → Template Palette)");
