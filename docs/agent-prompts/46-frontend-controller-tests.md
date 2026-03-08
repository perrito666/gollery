# Prompt 46 — Frontend controller tests

Add tests for frontend controllers.

Implement:
- tests for `AlbumController`: showRoot success, showAlbum success, 401/403/404 error handling
- tests for `AssetController`: showAsset success, error handling
- tests for `handleApiError` shared function
- tests for `Session`: restore, login, logout
- mock the API client with a fake that returns canned responses or throws `ApiError`

Do not add external test dependencies.
