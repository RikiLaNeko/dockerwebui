#/bin/sh

# Add all files to git 
git add . -p 

# Commit with the date and time of the commit 
git commit -m "Update on $(date) at $(date +%H:%M:%S)"

# Push to the remote repository
git push origin master


