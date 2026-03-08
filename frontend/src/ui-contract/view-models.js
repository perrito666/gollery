/**
 * View models define the data shape passed from controllers to views.
 *
 * Views receive view models, not raw API responses.
 */

/**
 * @typedef {Object} AlbumViewModel
 * @property {string} id
 * @property {string} title
 * @property {string} description
 * @property {string} path
 * @property {Array<{path: string}>} children
 * @property {Array<{id: string, filename: string, thumbnailURL: string}>} assets
 */

/**
 * @typedef {Object} AssetViewModel
 * @property {string} id
 * @property {string} filename
 * @property {string} albumPath
 * @property {string} albumId
 * @property {string} previewURL
 * @property {string} originalURL
 * @property {string|null} prevAssetId
 * @property {string|null} nextAssetId
 */

export default {};
