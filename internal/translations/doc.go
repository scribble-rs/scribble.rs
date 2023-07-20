/*
Package translations introduces a simple localization layer that can be
used by a templating engine.

Upon retrieving a localization package, values can be retrieved in both Go
code and the .html files via Get(key string). If a given key hasn't been
translated, the default localization pack will be accessed.

If the templating engine accesses non existent values, the server will panic.
This makes sure that we can't oversee the use of non-existent values.

Values must be plain text and not contain any HTML or CSS.
If values are to be inserted dynamically, placeholders can be placed with "%s".
*/
package translations
