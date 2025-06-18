# Zuper Connector Setup Guide

While developing the connector, please fill out this form. This information is needed to write docs and to help other users set up the connector.

---

## Connector capabilities

This connector syncs the following resources:

- Users
- Teams
- Roles
- Access Roles

It also supports provisioning for:

- Users
- Assigning and unassigning users to teams
- Updating a user's role

---

## Connector credentials

1. **What credentials or information are needed to set up the connector?**  
   This connector requires:  
   — Api URL  
   — Api key

   **Args**:  
   `--api-url`  
   `--api-key`

2. **For each item in the list above:**

   - **How does a user create or look up that credential or info?**

     1. Log in to [Zuper Pro](https://staging.zuperpro.com/login).
     2. Navigate to **Settings** → **Account Settings** → **API Keys**.
     3. Click on **New API Key**, enter a name for your key, and click **Generate**. The API key will be displayed.

    You will also need the **API URL** provided by Zuper.

---

