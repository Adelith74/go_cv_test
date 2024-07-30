let drop = document.getElementById("drop_file")
drop.addEventListener('change', () => {
    var list = document.getElementById("list")
    var name = drop.files[0].name
    addVideoToList(name)
})

function addVideoToList(video_name){
    var list = document.getElementById("list")
    var div = document.createElement("div")
    div.style = "display: flex; border-radius: 15px; border: 2px; border-color: black; border-style: solid"
    var p = document.createElement("p")
    p.textContent = video_name
    p.style = "color: black"
    var label = document.createElement("label")
    
    div.appendChild(p)
    list.appendChild(div)
}