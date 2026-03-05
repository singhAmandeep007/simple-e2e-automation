import React, { useState } from "react";

const FILE_ICON = "📄";
const FOLDER_ICON_OPEN = "📂";
const FOLDER_ICON_CLOSED = "📁";

/**
 * Builds a nested tree from a flat list of {path, isDir} entries.
 */
function buildTree(flat) {
  const root = { name: "/", children: {} };
  for (const entry of flat) {
    const parts = entry.path.split("/");
    let node = root;
    for (let i = 0; i < parts.length; i++) {
      const part = parts[i];
      if (!node.children[part]) {
        node.children[part] = {
          name: part,
          isDir: i < parts.length - 1 ? true : entry.isDir,
          size: entry.size,
          modTime: entry.modTime,
          children: {},
        };
      }
      node = node.children[part];
    }
  }
  return root;
}

function TreeNode({ name, node, depth = 0 }) {
  const [open, setOpen] = useState(depth < 2);
  const children = Object.values(node.children || {});
  const hasChildren = children.length > 0;

  return (
    <div style={{ paddingLeft: depth === 0 ? 0 : "20px" }}>
      <div
        data-testid={`tree-node-${name}`}
        style={{
          display: "flex",
          alignItems: "center",
          gap: "8px",
          padding: "4px 6px",
          borderRadius: "6px",
          cursor: node.isDir ? "pointer" : "default",
          color: node.isDir ? "var(--text)" : "var(--text-muted)",
          fontSize: "13px",
          userSelect: "none",
        }}
        onClick={() => node.isDir && setOpen((o) => !o)}
      >
        <span>{node.isDir ? (open && hasChildren ? FOLDER_ICON_OPEN : FOLDER_ICON_CLOSED) : FILE_ICON}</span>
        <span style={{ fontFamily: "monospace" }}>{name}</span>
        {!node.isDir && node.size != null && (
          <span style={{ fontSize: "11px", color: "#475569", marginLeft: "auto" }}>{formatSize(node.size)}</span>
        )}
      </div>
      {node.isDir && open && hasChildren && (
        <div>
          {children
            .sort((a, b) => b.isDir - a.isDir || a.name.localeCompare(b.name))
            .map((child) => (
              <TreeNode
                key={child.name}
                name={child.name}
                node={child}
                depth={depth + 1}
              />
            ))}
        </div>
      )}
    </div>
  );
}

function formatSize(bytes) {
  if (bytes < 1024) return `${bytes}B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)}KB`;
  return `${(bytes / 1024 / 1024).toFixed(1)}MB`;
}

/**
 * FolderTree renders a collapsible, deeply nested directory structure from a
 * flat array of file paths returned by the Go backend scan.
 *
 * @param {Object} props
 * @param {Array<{path: string, isDir: boolean, size: number, modTime: string}>} props.nodes - Flat array from SQLite `scan_tree`
 * @returns {JSX.Element}
 */
export default function FolderTree({ nodes }) {
  const tree = buildTree(nodes);
  const children = Object.values(tree.children);

  return (
    <div
      data-testid="folderTree"
      style={{ lineHeight: "1.6" }}
    >
      {children
        .sort((a, b) => b.isDir - a.isDir || a.name.localeCompare(b.name))
        .map((child) => (
          <TreeNode
            key={child.name}
            name={child.name}
            node={child}
            depth={0}
          />
        ))}
    </div>
  );
}
