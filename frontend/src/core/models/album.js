/**
 * Album model — core data representation.
 */

export class Album {
  constructor(data = {}) {
    this.id = data.id || '';
    this.title = data.title || '';
    this.path = data.path || '';
    this.children = data.children || [];
    this.assets = data.assets || [];
  }
}
