# Notes

## Use the file as a trigger

The `Persist` function will create a file at a specified path (with the TXT value and domain as content).
The file will be polled for changes.
A specific format is required.

The user will have to edit the file manually to add a specific value (ex: `"status": "ok"`).

## Use an HTTP call as a trigger

The `Persist` function will make an HTTP call.

I think it is not a secured approach.
