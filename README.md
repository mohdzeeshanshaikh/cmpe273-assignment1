#Stock Market
##How to use:
###To Buy:
####Request
go run client.go "GOOG:65%,YHOO:35%" 38500

**Format:** "stock1:percentage, stock2:percentage,..." budget
####Response 
Trade number

stock1 name:share number:stock1 price  stock2 name:share number:stock2 price ...

Uninvested Amount

###To Check:
####Request
go run client.go 101

**Format:** tradeId
####Response
stock1 name:share number:stock1 price  stock2 name:share number:stock2 price ...

Uninvested Amount

Total Market Value
