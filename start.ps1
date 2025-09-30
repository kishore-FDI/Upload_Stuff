# ASCII banner "Go Server"
$asciiBanner = @'
   ____       ____                          
  / ___| ___ / ___|  ___ _ ____   _____ _ __ 
 | |  _ / _ \ |  _ / _ \ '__\ \ / / _ \ '__|
 | |_| | (_) | |_| |  __/ |   \ V /  __/ |   
  \____|\___/ \____|\___|_|    \_/ \___|_|   
'@

# ASCII Gopher
$gopher = @'
     ,_---~~~~~----._
  _,,_,*^____      _____``*g*"*,
 / __/ /'     ^.  /      \ ^@q f
[  @f | @))    |  | @))   l  0 _/
 \`/   \~____ / __ \~___/    \
  |           _l__l_           I
  }          [______]           I
  ]            | | |            |
  ]             ~ ~             |
  |                            |
   \                          /
'@

# Print in Cyan
Write-Host $asciiBanner -ForegroundColor Cyan
# Write-Host $gopher -ForegroundColor Cyan
Write-Host "Starting the Go Server..." -ForegroundColor Cyan

# Run the Go server
go run main.go
