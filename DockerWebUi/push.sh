#/bin/sh

# Add all files to git 
git add .

# Commit with the date and time of the commit 
git commit -m "Update on $(date)"

# Push to the remote repository
git push master master


