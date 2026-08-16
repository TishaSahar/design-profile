# Requirements:

1. The administrator (designer) must log in via email using a one-time password sent to their inbox. (The application’s account must use the Gmail API for sending emails; the application’s email settings must be hidden from the user).

2. The designer can upload new photos of their projects via the administrator page UI. Media files can be stored as binary files in a PostgreSQL table, the schema for which is described in the ./backend/migrations directory. When uploading a new project, the designer can add a text description of up to 500 characters and select a cover photo for the project.

3. The designer/administrator must have convenient access to the database of contacts and project requests.

4. The client or website user must be able to view the designer’s portfolio, featuring cover photos of all projects, and have the option to examine each project in more detail by clicking on the cover photo or description.

5. The client must be able to view the designer’s contact details (links to Telegram, Instagram and email address).

6. The client can submit a project request. They should provide their first and last name, contact details, a project description and attach up to 10 photos or a single PDF file containing project drawings.

7. When creating a project request, the client must be able to view the personal data transfer agreement and must consent to it before saving the project request.




Translated with DeepL.com (free version)