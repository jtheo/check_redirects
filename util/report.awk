BEGIN {
	FS=","
}
{
	if ($3 != 200)
		errors = errors + 1
	if ($4 > 1 )
		redir = redir + 1
}

END {
	diff = total-NR
	percDone = NR * 100 / total 
	percToDo = diff * 100 / total
	printf "Processed: %d (%d%%), Remaining: %d (%d%%), Errors: %d, Too Many redirects (> 1): %d\n",NR,percDone,diff,percToDo, errors,redir
}
