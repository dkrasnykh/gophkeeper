/*
cli module provides user UI commands.

# Selection commands
The selection commands include package models:

	view_auth model prompts to select an action from the list {"Login", "Register"}.
	view_command_list model prompts to select an action from the list {"Get all secrets", "Add credentials", "Add text data", "Add binary data", "Add card data"}

# Register

	view_register model provides form for indicate login, password. It includes widget for data submission.

# Login

	view_login model provides form for indicate login, password. It includes widget for data submission.

# Get all secrets

	view_list model uses for show all private user data.

# Add credentials

	view_add_credentials model provides form for indicate tag, login, password, comment. It includes widget for data submission.

# Add text data

	view_add_text model provides form for indicate tag, key, value, comment. It includes widget for data submission.

# Add binary data

	view_add_binary model provides form for indicate tag, key, value, comment. It includes widget for data submission.

# Add card data

	view_add_card model provides form for indicate tag, number, exp, cvv, comment. It includes widget for data submission.
*/
package cli
