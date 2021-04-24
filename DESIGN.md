# THE DESIGN

With temporary name bssh

Possible commands 
add // add using ip ranges
delete
list // possbly with namespaces or selectors
run // can be used to run multiple commands in a remote or multipl using ip ranges(runed only if in the list) or in name like *.google.com or best using namespaces like -n development  
rand

$ bssh
prints the quick select screen
interactive
however this requires work. For now use a simple list and select from there


bssh add ssh foucault@8.8.8.8 
enter the password or the path for ssh key
(password can be in the form of a path so make sure to validate if path exists and also valid)
also promting is never good. So -p or -k is also should be options

bssh add asd@asd
bssh run *@asd -- date
bssh rand -n development
bssh rand -n development master


NEW DESING
Since we establised the fzf interface we can simplitfy the interaction

bssh
Shows the list where you can select things
Side window shows all metadata that used by it, except password.
Side window also shouws the ssh command that will be exected on selection
if you -p it, it will show passwords too
NOTE: For now password is kept as plani text.

bssh -n 
opens up the window for namespace selection. 
after namespace selection, bssh opens a new window for ssh selection
if you -p it, it will show passwords too

bssh -n <namespace>
pre select namespace

bssh add -n <namespace> furkan@192.168.1.1 --password hello_world # or -p 
adds this ssh connection to the bucket 
if -n is not specified it is added to default

bssh add -n <namespace> furkan@192.168.1.1 --key /path/to/the/ssh/key # or -k 
adds this ssh connection to the bucket 
if -n is not specified it is added to default

bssh -r -n <namepsace>
remove connection
if namespace is not specified bssh searches for all namespaces

bssh list 
list all connections but dont use fzf
again you can optionaly add -n for namespace








