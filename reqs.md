## Nyusu ui requirements

## Login/Register Page

- login with password
- register with password, double password to correctly register
- message when auth credentials are wrong
- message when the email is already registered

## Home Page

#### List of posts

- screen/message when there are no posts
- shows recent most recent posts with pagination
- you can see the title, an optional (not really) description
- when you click a post it should send you to the page to read it
- should have a way to refresh the feed (maybe)
- you can "bookmark"/"like" a post
- you can remove the bookmark/like to a post
- you can click a button to add a new feed

#### Filter options

- can filter by list of feeds that the user follow (can be a dropdown, maybe)
- can filter by liked posts

## New Feed popup

- should have an input with autocomplete for feeds that exist
- if feed does not exists, should let you add an url to the rss feed
- validation in case it is an invalid feed
- loading state when it's adding it, since it tries to fetch the posts
- should take you back to homepage on success or close

## Liked posts Page

- same as home page but only bookmarked posts
- also has pagination
- can order by date you liked it or date of the post
