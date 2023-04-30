
function doRequest(url, request, goodLocation) {
    fetch(url, request)
        .then(response => {
            response.text().then(text => {
                
                    document.getElementById("result").textContent = text               
            })

            if (!response.ok) {
                if (response.status == 401) {
                    window.location.href = "login.html"
                }
                return
            }

            if (goodLocation != null) {
                window.location.href = goodLocation
            }
        })
        .catch(error => {
            console.error(error.message);
            document.getElementById("result").textContent = error.message;
        });
}



function submitLoginForm(event) {
    event.preventDefault();
    const loginData = {
        username: document.getElementById("username").value,
        password: document.getElementById("password").value
    };
    doRequest("/login", {
        method: "POST",
        headers: {
            "Content-Type": "application/json"
        },
        body: JSON.stringify(loginData)
    }, "/confirm.html")
}

