var forms = (function () {
    const contactFrom = document.forms["contact"]
    if (contactFrom){
        contactFrom.addEventListener("submit", e => {
            e.preventDefault()
            alert("contact form submitted")
        })
    }
})()