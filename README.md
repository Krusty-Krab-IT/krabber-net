# Krabber.Net

Mr. Krabs had the Krusty Krab R&D  develop a new way to be dumb with blogging thanks to a grant from the Bikini Bottom Modernization effort.

It is an open-source project (we told Mr.Krabs we knew of a way to get free labor, so he naturally agreed), feel free to 
fork it, clone it, and deploy it in your own AWS account or completely refactor the database (ouch) to be something else.

## Service Level Agreement (SLA)

The Bikini Bottom Modernization effort grant covers the AWS cloud bill of $100 USD per month. 
After, that the servers are turned off and the site will be down, so if you're visiting the site and it isn't working 
then that is probably why.

Mr.Krabs isn't going to pay out of his own pocket after all...

## Pull Request Ideas


- [ ] Do a complete rewrite in Rust for better alignment with brand
- [x] Page refresh after form submission (molts...)
- [x] Name in thread should show crab name that made it
- [x] thread <- back button should work
- [ ] deploy api endpoint for website...
- [x] remolt button should work (backend exists)
- [x] like button should work (backend exists)
- [ ] create password reset page (backend exists)
- [ ] create mr.krabs bot
- [ ] URL shortener with hyperlink 
- [ ] view count on molts
- [ ] Notifications
  - [ ] '@' mentions i.f.f implement the complete feature
  - [x] New Follower
  - [x] Like
  - [x] Remolt
  - [x] Comment -> backwards 

- [ ] Settings
- [ ] Refactor HTML templates to not be so repetitive 
- [ ] Make activate page on brand
- [x] See who to follow (link to all crabs)
- [ ] /sea Create SEA Lambda 
- [ ] /crabs make it same colors as others & general layout
- [ ] /molts 



## Single Table Design

Rearchitecting twitter as a server side rendered web app with a single table was the goal of this as an exercise, so 
if you're curious about what that looks like then:
- Check DOCS/provisioun_table.py for the script to create the table in your AWS Account
- Check DOCS/table.json to see what the schema is if you're curious 




## For Funsies

If you're new to programming with GoLang and are looking for some projects to add to your resume then consider refactoring 
the architecture of the application to:

- [ ] Single Page Application
- [ ] Micro Service Based
- [ ] Serverless based
- [ ] Rewrite it all in Rust
